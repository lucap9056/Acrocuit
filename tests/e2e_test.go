package test

import (
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type client struct {
	t             *testing.T
	http          *http.Client
	base          *url.URL
	identityToken string
	accessToken   string
}

func newClient(t *testing.T) *client {
	t.Helper()
	base, err := url.Parse(host + apiPath)
	require.NoError(t, err, "parse base url")
	c := &client{
		t:             t,
		http:          &http.Client{Timeout: 10 * time.Second},
		base:          base,
		identityToken: identityToken,
		accessToken:   env("ACROCUIT_ACCESS_TOKEN", ""),
	}
	c.requireServer()
	return c
}

func (c *client) requireServer() {
	c.t.Helper()
	res, err := c.http.Get(c.base.String())
	if err != nil {
		c.t.Skipf("server unreachable at %s: %v", c.base, err)
	}
	res.Body.Close()
}

func (c *client) with(t *testing.T) *client {
	clone := *c
	clone.t = t
	return &clone
}

func (c *client) do(req *http.Request) (int, http.Header, []byte) {
	c.t.Helper()

	req.URL = c.base.JoinPath(req.URL.Path)
	if c.identityToken != "" {
		req.Header.Set("X-Forwarded-Identity", c.identityToken)
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	res, err := c.http.Do(req)
	require.NoError(c.t, err, "%s %s", req.Method, req.URL.Path)
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	require.NoError(c.t, err, "%s %s: read body", req.Method, req.URL.Path)
	if res.StatusCode >= http.StatusBadRequest {
		c.t.Logf("%s %s -> %d %s", req.Method, req.URL.Path, res.StatusCode, raw)
	}
	return res.StatusCode, res.Header, raw
}

func (c *client) Send(req *http.Request, data any) int {
	c.t.Helper()

	status, _, raw := c.do(req)
	if data != nil && len(raw) > 0 {
		err := json.Unmarshal(raw, &response.Response[any]{Data: data})
		require.NoError(c.t, err, "%s %s: decode response (%s)", req.Method, req.URL.Path, raw)
	}
	return status
}

func (c *client) BackgroundImage(spaceId int, etag string) (int, http.Header, []byte) {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/spaces/%d/background-image", spaceId), nil)
	require.NoError(c.t, err, "new request")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	return c.do(req)
}

func newRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err, "marshal body")
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, path, reader)
	require.NoError(t, err, "new request")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func downstream(c *client, switchId int) ([]models.Switch, []models.Device) {
	c.t.Helper()

	path := fmt.Sprintf("/switches/%d/downstream", switchId)
	var tree map[string]json.RawMessage
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &tree)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)

	var switches []models.Switch
	var devices []models.Device
	require.NoError(c.t, json.Unmarshal(tree["switches"], &switches), "decode downstream switches")
	require.NoError(c.t, json.Unmarshal(tree["devices"], &devices), "decode downstream devices")
	return switches, devices
}

func upstream(c *client, deviceId int) []models.Switch {
	c.t.Helper()

	path := fmt.Sprintf("/devices/%d/upstream", deviceId)
	var switches []models.Switch
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &switches)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)
	return switches
}

func idsOf(ids map[string]int, names ...string) []int {
	result := make([]int, len(names))
	for i, name := range names {
		result[i] = ids[name]
	}
	return result
}

func switchIds(switches []models.Switch) []int {
	result := make([]int, len(switches))
	for i, b := range switches {
		result[i] = b.Id
	}
	return result
}

func hasDeviceOnSwitch(devices []models.Device, deviceId, switchId int) bool {
	return slices.ContainsFunc(devices, func(d models.Device) bool {
		return d.Id == deviceId && d.Switch != nil && d.Switch.Id == switchId
	})
}

func logTree(t *testing.T, rootId int, switches []models.Switch, devices []models.Device) {
	t.Helper()

	children := make(map[int][]models.Switch)
	names := make(map[int]string, len(switches))
	for _, b := range switches {
		names[b.Id] = b.Name
		if b.UpstreamSwitch != nil {
			children[b.UpstreamSwitch.Id] = append(children[b.UpstreamSwitch.Id], b)
		}
	}
	attached := make(map[int][]models.Device)
	for _, d := range devices {
		if d.Switch != nil {
			attached[d.Switch.Id] = append(attached[d.Switch.Id], d)
		}
	}

	var sb strings.Builder
	var walk func(id int, prefix string, last bool)
	walk = func(id int, prefix string, last bool) {
		branch, indent := "├─ ", "│  "
		if last {
			branch, indent = "└─ ", "   "
		}
		fmt.Fprintf(&sb, "%s%s[B] %s (#%d)\n", prefix, branch, names[id], id)

		kids, devs := children[id], attached[id]
		for i, d := range devs {
			leaf := "├─ "
			if i == len(devs)-1 && len(kids) == 0 {
				leaf = "└─ "
			}
			fmt.Fprintf(&sb, "%s%s[D] %s (#%d)\n", prefix+indent, leaf, d.Name, d.Id)
		}
		for i, k := range kids {
			walk(k.Id, prefix+indent, i == len(kids)-1)
		}
	}
	walk(rootId, "", true)
	t.Logf("downstream of #%d:\n%s", rootId, sb.String())
}

func pngImage(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size)))
	require.NoError(t, err, "encode png")
	return buf.Bytes()
}

func backgroundImageRequest(t *testing.T, spaceId int, field string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, "background")
	require.NoError(t, err, "create form file")
	part.Write(data)
	writer.Close()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/spaces/%d/background-image", spaceId), &body)
	require.NoError(t, err, "new request")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestE2E(t *testing.T) {
	c := newClient(t)

	body := map[string]any{"name": "space group"}

	var spaceGroup models.SpaceGroup
	req := newRequest(t, http.MethodPost, "/space-groups", body)
	status := c.Send(req, &spaceGroup)
	require.Equal(t, http.StatusOK, status, "create space group")
	spaceGroupPath := fmt.Sprintf("/space-groups/%d", spaceGroup.Id)
	t.Cleanup(func() {
		req := newRequest(t, http.MethodDelete, spaceGroupPath, nil)
		c.Send(req, nil)
	})

	spaces := make(map[string]int)
	for _, name := range []string{"space1", "space2", "space3", "space4", "space5"} {
		path := spaceGroupPath + "/spaces"
		body := map[string]any{"name": name}

		var space models.Space
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &space)
		require.Equal(t, http.StatusOK, status, "create space %s", name)
		spaces[name] = space.Id
	}

	switchPanels := make(map[string]int)
	for _, tc := range []struct{ space, name string }{
		{"space1", "group1"},
		{"space2", "group2"},
		{"space3", "group3"},
		{"space4", "group4"},
		{"space5", "group5"},
	} {
		path := fmt.Sprintf("/spaces/%d/switch-panels", spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var switchPanel models.SwitchPanel
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &switchPanel)
		require.Equal(t, http.StatusOK, status, "create switch panel %s", tc.name)
		switchPanels[tc.name] = switchPanel.Id
	}

	switches := make(map[string]int)
	for _, tc := range []struct{ group, name, upstream string }{
		{"group1", "switch1", ""},
		{"group1", "switch2", "switch1"},
		{"group1", "switch3", "switch1"},
		{"group2", "switch4", "switch2"},
		{"group2", "switch5", "switch2"},
		{"group3", "switch6", "switch2"},
		{"group4", "switch7", "switch3"},
		{"group5", "switch8", "switch3"},
		{"group2", "switch9", "switch5"},
		{"group4", "switch10", "switch7"},
		{"group5", "switch11", "switch8"},
	} {
		path := fmt.Sprintf("/switch-panels/%d/switches", switchPanels[tc.group])
		body := map[string]any{"name": tc.name}
		if tc.upstream != "" {
			body["upstream_switch_id"] = switches[tc.upstream]
		}

		var sw models.Switch
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &sw)
		require.Equal(t, http.StatusOK, status, "create switch %s", tc.name)
		switches[tc.name] = sw.Id
	}

	root := switches["switch1"]
	left := switches["switch10"]
	right := switches["switch11"]

	t.Run("SwitchUpstreamRejectsCycle", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct{ sw, upstream string }{
			{"switch1", "switch10"},
			{"switch10", "switch10"},
		} {
			path := fmt.Sprintf("/switches/%d/upstream", switches[tc.sw])
			body := map[string]any{"upstream_switch_id": switches[tc.upstream]}

			req := newRequest(t, http.MethodPatch, path, body)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusBadRequest, status, "set upstream of %s to %s", tc.sw, tc.upstream)
		}
	})

	devices := make(map[string]int)
	for _, tc := range []struct{ space, name string }{
		{"space2", "device1"},
		{"space2", "device2"},
		{"space2", "device3"},
		{"space2", "device4"},
		{"space3", "device5"},
		{"space3", "device6"},
		{"space3", "device7"},
		{"space4", "device8"},
		{"space4", "device9"},
		{"space5", "device10"},
		{"space5", "device11"},
	} {
		path := fmt.Sprintf("/spaces/%d/devices", spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var device models.Device
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &device)
		require.Equal(t, http.StatusOK, status, "create device %s", tc.name)
		devices[tc.name] = device.Id
	}

	shared := devices["device9"]

	for _, tc := range []struct{ device, sw string }{
		{"device1", "switch4"},
		{"device2", "switch4"},
		{"device3", "switch5"},
		{"device4", "switch9"},
		{"device5", "switch6"},
		{"device6", "switch6"},
		{"device7", "switch6"},
		{"device8", "switch7"},
		{"device9", "switch10"},
		{"device9", "switch11"},
		{"device10", "switch8"},
		{"device11", "switch8"},
	} {
		path := fmt.Sprintf("/devices/%d/switches", devices[tc.device])
		body := map[string]any{"switch_id": switches[tc.sw]}

		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "attach %s to %s", tc.device, tc.sw)
	}

	t.Run("AttachDeviceRejectsDuplicate", func(t *testing.T) {
		c := c.with(t)
		path := fmt.Sprintf("/devices/%d/switches", shared)
		body := map[string]any{"switch_id": left}

		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, nil)
		assert.Equal(t, http.StatusConflict, status, "attach device9 to switch10 again")
	})

	allSwitches := make([]string, 0, len(switches))
	for name := range switches {
		allSwitches = append(allSwitches, name)
	}

	t.Run("SwitchDownstreamIncludesDescendants", func(t *testing.T) {
		c := c.with(t)
		treeSwitches, treeDevices := downstream(c, root)
		logTree(t, root, treeSwitches, treeDevices)

		assert.ElementsMatch(t, idsOf(switches, allSwitches...), switchIds(treeSwitches), "downstream switches")
		assert.True(t, hasDeviceOnSwitch(treeDevices, shared, left), "device9 should appear under switch10")
		assert.True(t, hasDeviceOnSwitch(treeDevices, shared, right), "device9 should appear under switch11")
	})

	t.Run("DeviceUpstreamDeduplicatesSharedPath", func(t *testing.T) {
		c := c.with(t)
		want := idsOf(switches, "switch10", "switch7", "switch11", "switch8", "switch3", "switch1")
		assert.ElementsMatch(t, want, switchIds(upstream(c, shared)), "upstream of device9")
	})

	spaceId := spaces["space2"]
	spacePath := fmt.Sprintf("/spaces/%d", spaceId)
	backgroundImagePath := spacePath + "/background-image"

	t.Run("SpaceBackgroundImageRoundTrip", func(t *testing.T) {
		c := c.with(t)
		status, _, _ := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusNotFound, status, "GET background image before upload")

		var space models.Space
		req := newRequest(t, http.MethodGet, spacePath, nil)
		status = c.Send(req, &space)
		require.Equal(t, http.StatusOK, status, "GET %s", spacePath)
		createdAt := space.BackgroundImageUpdatedAt

		time.Sleep(5 * time.Millisecond)
		first := pngImage(t, 2)
		var uploaded models.Space
		req = backgroundImageRequest(t, spaceId, "image", first)
		status = c.Send(req, &uploaded)
		require.Equal(t, http.StatusOK, status, "upload background image")
		uploadedAt := uploaded.BackgroundImageUpdatedAt
		assert.Greater(t, uploadedAt, createdAt, "background_image_updated_at after upload")

		status, header, got := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusOK, status, "GET background image")
		assert.Equal(t, "image/png", header.Get("Content-Type"), "GET background image Content-Type")
		assert.Equal(t, first, got, "GET background image body")

		etag := header.Get("ETag")
		assert.Equal(t, fmt.Sprintf(`"%d"`, uploadedAt), etag, "ETag")
		status, _, _ = c.BackgroundImage(spaceId, etag)
		assert.Equal(t, http.StatusNotModified, status, "GET with matching If-None-Match")

		second := pngImage(t, 4)
		req = backgroundImageRequest(t, spaceId, "image", second)
		status = c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "replace background image")
		_, _, got = c.BackgroundImage(spaceId, "")
		assert.Equal(t, second, got, "GET after replace")

		for _, tc := range []struct {
			method string
			want   int
		}{
			{http.MethodDelete, http.StatusOK},
			{http.MethodGet, http.StatusNotFound},
			{http.MethodDelete, http.StatusNotFound},
		} {
			req := newRequest(t, tc.method, backgroundImagePath, nil)
			status := c.Send(req, nil)
			assert.Equal(t, tc.want, status, "%s background image", tc.method)
		}
	})

	t.Run("SpaceBackgroundImageRejectsInvalidUpload", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct {
			space int
			field string
			data  []byte
			want  int
		}{
			{spaceId, "image", []byte("plain text, not an image"), http.StatusUnsupportedMediaType},
			{spaceId, "file", pngImage(t, 2), http.StatusBadRequest},
			{spaceId, "image", make([]byte, 11<<20), http.StatusRequestEntityTooLarge},
			{0, "image", pngImage(t, 2), http.StatusNotFound},
		} {
			req := backgroundImageRequest(t, tc.space, tc.field, tc.data)
			status := c.Send(req, nil)
			assert.Equal(t, tc.want, status, "upload %q (%d bytes) to space %d", tc.field, len(tc.data), tc.space)
		}
		status, _, _ := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusNotFound, status, "GET background image after rejected uploads")
	})

	t.Run("DeleteSpaceRemovesCrossSpaceReferences", func(t *testing.T) {
		c := c.with(t)
		removedPath := fmt.Sprintf("/spaces/%d", spaces["space5"])
		req := newRequest(t, http.MethodDelete, removedPath, nil)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "delete space5")

		for _, path := range []string{
			removedPath,
			fmt.Sprintf("/switch-panels/%d", switchPanels["group5"]),
			fmt.Sprintf("/switches/%d", switches["switch8"]),
			fmt.Sprintf("/switches/%d", right),
			fmt.Sprintf("/devices/%d", devices["device10"]),
			fmt.Sprintf("/devices/%d", devices["device11"]),
		} {
			req := newRequest(t, http.MethodGet, path, nil)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusNotFound, status, "GET %s after deleting space5", path)
		}

		want := idsOf(switches, "switch10", "switch7", "switch3", "switch1")
		assert.ElementsMatch(t, want, switchIds(upstream(c, shared)), "upstream of device9 after deleting space5")

		remaining := slices.DeleteFunc(slices.Clone(allSwitches), func(name string) bool {
			return name == "switch8" || name == "switch11"
		})
		treeSwitches, treeDevices := downstream(c, root)
		logTree(t, root, treeSwitches, treeDevices)

		assert.ElementsMatch(t, idsOf(switches, remaining...), switchIds(treeSwitches), "downstream switches after deleting space5")
		assert.True(t, hasDeviceOnSwitch(treeDevices, shared, left), "device9 should remain under switch10")
		assert.True(t, hasDeviceOnSwitch(treeDevices, devices["device3"], switches["switch5"]), "device3 should be unaffected")
	})

	t.Run("DeleteSpaceGroupCascades", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodDelete, spaceGroupPath, nil)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "delete space group")

		req = newRequest(t, http.MethodGet, spaceGroupPath, nil)
		status = c.Send(req, nil)
		assert.Equal(t, http.StatusNotFound, status, "GET %s after delete", spaceGroupPath)
		for _, tc := range []struct {
			prefix string
			ids    map[string]int
		}{
			{"/spaces", spaces},
			{"/switch-panels", switchPanels},
			{"/switches", switches},
			{"/devices", devices},
		} {
			for _, id := range tc.ids {
				path := fmt.Sprintf("%s/%d", tc.prefix, id)
				req := newRequest(t, http.MethodGet, path, nil)
				status := c.Send(req, nil)
				assert.Equal(t, http.StatusNotFound, status, "GET %s after deleting space group", path)
			}
		}
	})
}
