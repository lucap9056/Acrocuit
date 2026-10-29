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

	"github.com/google/uuid"
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

	target := c.base.JoinPath(req.URL.Path)
	target.RawQuery = req.URL.RawQuery
	req.URL = target
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

func (c *client) BackgroundImage(spaceId string, etag string) (int, http.Header, []byte) {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/spaces/%s/background-image", spaceId), nil)
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

func powerChain(c *client, path string) models.PowerChain {
	c.t.Helper()

	var chain models.PowerChain
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &chain)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)
	return chain
}

func downstream(c *client, switchId string) models.PowerChain {
	c.t.Helper()
	return powerChain(c, fmt.Sprintf("/switches/%s/downstream", switchId))
}

func upstream(c *client, deviceId string) models.PowerChain {
	c.t.Helper()
	return powerChain(c, fmt.Sprintf("/devices/%s/upstream", deviceId))
}

func idsOf(ids map[string]string, names ...string) []string {
	result := make([]string, len(names))
	for i, name := range names {
		result[i] = ids[name]
	}
	return result
}

func switchIds(switches []models.Switch) []string {
	result := make([]string, len(switches))
	for i, b := range switches {
		result[i] = string(b.Id)
	}
	return result
}

func chainSwitchIds(switches []models.ChainSwitch) []string {
	result := make([]string, len(switches))
	for i, b := range switches {
		result[i] = string(b.Id)
	}
	return result
}

func hasDeviceOnSwitch(chain models.PowerChain, deviceId, switchId string) bool {
	return slices.ContainsFunc(chain.DeviceLinks, func(l models.DeviceLink) bool {
		return string(l.DeviceId) == deviceId && string(l.SwitchId) == switchId
	})
}

func findSwitchLink(chain models.PowerChain, upstreamId, downstreamId string) (models.SwitchLink, bool) {
	index := slices.IndexFunc(chain.SwitchLinks, func(l models.SwitchLink) bool {
		return string(l.UpstreamSwitchId) == upstreamId && string(l.DownstreamSwitchId) == downstreamId
	})
	if index < 0 {
		return models.SwitchLink{}, false
	}
	return chain.SwitchLinks[index], true
}

func logTree(t *testing.T, rootId string, chain models.PowerChain) {
	t.Helper()

	names := make(map[string]string, len(chain.Switches))
	for _, b := range chain.Switches {
		names[string(b.Id)] = b.Name
	}
	children := make(map[string][]string)
	for _, l := range chain.SwitchLinks {
		children[string(l.UpstreamSwitchId)] = append(children[string(l.UpstreamSwitchId)], string(l.DownstreamSwitchId))
	}
	devices := make(map[string]models.Device, len(chain.Devices))
	for _, d := range chain.Devices {
		devices[string(d.Id)] = d
	}
	attached := make(map[string][]models.Device)
	for _, l := range chain.DeviceLinks {
		attached[string(l.SwitchId)] = append(attached[string(l.SwitchId)], devices[string(l.DeviceId)])
	}

	var sb strings.Builder
	var walk func(id string, prefix string, last bool)
	walk = func(id string, prefix string, last bool) {
		branch, indent := "├─ ", "│  "
		if last {
			branch, indent = "└─ ", "   "
		}
		fmt.Fprintf(&sb, "%s%s[B] %s (#%s)\n", prefix, branch, names[id], id)

		kids, devs := children[id], attached[id]
		for i, d := range devs {
			leaf := "├─ "
			if i == len(devs)-1 && len(kids) == 0 {
				leaf = "└─ "
			}
			fmt.Fprintf(&sb, "%s%s[D] %s (#%s)\n", prefix+indent, leaf, d.Name, d.Id)
		}
		for i, k := range kids {
			walk(k, prefix+indent, i == len(kids)-1)
		}
	}
	walk(rootId, "", true)
	t.Logf("downstream of #%s:\n%s", rootId, sb.String())
}

func pngImage(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size)))
	require.NoError(t, err, "encode png")
	return buf.Bytes()
}

func backgroundImageRequest(t *testing.T, spaceId string, field string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, "background")
	require.NoError(t, err, "create form file")
	part.Write(data)
	writer.Close()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/spaces/%s/background-image", spaceId), &body)
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
	spaceGroupPath := fmt.Sprintf("/space-groups/%s", spaceGroup.Id)
	t.Cleanup(func() {
		req := newRequest(t, http.MethodDelete, spaceGroupPath, nil)
		c.Send(req, nil)
	})

	spaces := make(map[string]string)
	for _, name := range []string{"space1", "space2", "space3", "space4", "space5"} {
		path := spaceGroupPath + "/spaces"
		body := map[string]any{"name": name}

		var space models.Space
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &space)
		require.Equal(t, http.StatusOK, status, "create space %s", name)
		spaces[name] = string(space.Id)
	}

	switchPanels := make(map[string]string)
	for _, tc := range []struct{ space, name string }{
		{"space1", "group1"},
		{"space2", "group2"},
		{"space3", "group3"},
		{"space4", "group4"},
		{"space5", "group5"},
	} {
		path := fmt.Sprintf("/spaces/%s/switch-panels", spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var switchPanel models.SwitchPanel
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &switchPanel)
		require.Equal(t, http.StatusOK, status, "create switch panel %s", tc.name)
		switchPanels[tc.name] = string(switchPanel.Id)
	}

	switches := make(map[string]string)
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
		path := fmt.Sprintf("/switch-panels/%s/switches", switchPanels[tc.group])
		body := map[string]any{"name": tc.name}

		var sw models.Switch
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &sw)
		require.Equal(t, http.StatusOK, status, "create switch %s", tc.name)
		switches[tc.name] = string(sw.Id)

		if tc.upstream != "" {
			path := fmt.Sprintf("/switches/%s/upstreams", sw.Id)
			body := map[string]any{"upstream_switch_id": switches[tc.upstream], "attributes": map[string]any{"from": tc.upstream}}
			req := newRequest(t, http.MethodPost, path, body)
			status := c.Send(req, nil)
			require.Equal(t, http.StatusOK, status, "link %s to %s", tc.name, tc.upstream)
		}
	}

	secondUpstreamPath := fmt.Sprintf("/switches/%s/upstreams", switches["switch9"])
	secondUpstream := map[string]any{"upstream_switch_id": switches["switch6"]}
	req = newRequest(t, http.MethodPost, secondUpstreamPath, secondUpstream)
	require.Equal(t, http.StatusOK, c.Send(req, nil), "link switch9 to switch6")

	root := switches["switch1"]
	left := switches["switch10"]
	right := switches["switch11"]

	t.Run("SwitchUpstreamRejectsCycle", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct{ sw, upstream string }{
			{"switch1", "switch10"},
			{"switch3", "switch11"},
			{"switch10", "switch10"},
		} {
			path := fmt.Sprintf("/switches/%s/upstreams", switches[tc.sw])
			body := map[string]any{"upstream_switch_id": switches[tc.upstream]}

			req := newRequest(t, http.MethodPost, path, body)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusBadRequest, status, "add upstream %s to %s", tc.upstream, tc.sw)
		}
	})

	t.Run("IncludeRejectsUnknownColumn", func(t *testing.T) {
		c := c.with(t)
		for _, path := range []string{
			spaceGroupPath,
			fmt.Sprintf("/spaces/%s", spaces["space1"]),
			fmt.Sprintf("/spaces/%s/switch-panels", spaces["space1"]),
			fmt.Sprintf("/spaces/%s/devices", spaces["space1"]),
			fmt.Sprintf("/switches/%s", root),
		} {
			req := newRequest(t, http.MethodGet, path+"?include=id,unknown", nil)
			assert.Equal(t, http.StatusBadRequest, c.Send(req, nil), "GET %s", path)
		}
	})

	t.Run("SwitchAcceptsMultipleUpstreams", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodPost, secondUpstreamPath, secondUpstream)
		assert.Equal(t, http.StatusConflict, c.Send(req, nil), "link switch9 to switch6 again")

		var upstreams []models.Switch
		req = newRequest(t, http.MethodGet, secondUpstreamPath, nil)
		require.Equal(t, http.StatusOK, c.Send(req, &upstreams), "GET %s", secondUpstreamPath)
		assert.ElementsMatch(t, idsOf(switches, "switch5", "switch6"), switchIds(upstreams), "upstreams of switch9")
	})

	devices := make(map[string]string)
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
		path := fmt.Sprintf("/spaces/%s/devices", spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var device models.Device
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &device)
		require.Equal(t, http.StatusOK, status, "create device %s", tc.name)
		devices[tc.name] = string(device.Id)
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
		path := fmt.Sprintf("/devices/%s/switches", devices[tc.device])
		body := map[string]any{"switch_id": switches[tc.sw]}

		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "attach %s to %s", tc.device, tc.sw)
	}

	t.Run("AttachDeviceRejectsDuplicate", func(t *testing.T) {
		c := c.with(t)
		path := fmt.Sprintf("/devices/%s/switches", shared)
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
		chain := downstream(c, root)
		logTree(t, root, chain)

		assert.ElementsMatch(t, idsOf(switches, allSwitches...), chainSwitchIds(chain.Switches), "downstream switches")
		for _, l := range chain.SwitchLinks {
			assert.Contains(t, chainSwitchIds(chain.Switches), l.UpstreamSwitchId, "downstream link upstream end")
			assert.Contains(t, chainSwitchIds(chain.Switches), l.DownstreamSwitchId, "downstream link downstream end")
		}
		assert.Len(t, chain.Devices, len(devices), "downstream devices")
		assert.True(t, hasDeviceOnSwitch(chain, shared, left), "device9 should appear under switch10")
		assert.True(t, hasDeviceOnSwitch(chain, shared, right), "device9 should appear under switch11")

		link, ok := findSwitchLink(chain, switches["switch2"], switches["switch4"])
		if assert.True(t, ok, "link switch2 -> switch4") {
			assert.JSONEq(t, `{"from":"switch2"}`, string(link.Attributes), "link switch2 -> switch4 attributes")
		}
	})

	t.Run("DeviceUpstreamFollowsEveryUpstream", func(t *testing.T) {
		c := c.with(t)
		chain := upstream(c, devices["device4"])
		want := idsOf(switches, "switch9", "switch5", "switch6", "switch2", "switch1")
		assert.ElementsMatch(t, want, chainSwitchIds(chain.Switches), "upstream of device4")
		_, ok := findSwitchLink(chain, switches["switch6"], switches["switch9"])
		assert.True(t, ok, "link switch6 -> switch9")
	})

	t.Run("DeviceUpstreamDeduplicatesSharedPath", func(t *testing.T) {
		c := c.with(t)
		want := idsOf(switches, "switch10", "switch7", "switch11", "switch8", "switch3", "switch1")
		chain := upstream(c, shared)
		assert.ElementsMatch(t, want, chainSwitchIds(chain.Switches), "upstream of device9")
		assert.Len(t, chain.SwitchLinks, len(want)-1, "upstream switch links of device9")
		linked := make([]string, len(chain.DeviceLinks))
		for i, l := range chain.DeviceLinks {
			linked[i] = string(l.SwitchId)
		}
		assert.ElementsMatch(t, []string{left, right}, linked, "device links of device9")
	})

	t.Run("PowerChainIncludesPanelsAndPositions", func(t *testing.T) {
		c := c.with(t)
		panelPosition := models.Position{X: 1, Y: 2, Z: 3}
		devicePosition := models.Position{X: 4, Y: 5, Z: 6}

		path := fmt.Sprintf("/switch-panels/%s/position", switchPanels["group4"])
		req := newRequest(t, http.MethodPut, path, map[string]any{"position": panelPosition})
		require.Equal(t, http.StatusOK, c.Send(req, nil), "PUT %s", path)

		path = fmt.Sprintf("/devices/%s/position", devices["device8"])
		req = newRequest(t, http.MethodPut, path, map[string]any{"position": devicePosition})
		require.Equal(t, http.StatusOK, c.Send(req, nil), "PUT %s", path)

		assertPanel := func(sw models.ChainSwitch, source string) {
			t.Helper()
			panel := sw.SwitchPanel
			require.NotNil(t, panel, "%s: switch #%s panel", source, sw.Id)
			require.NotNil(t, panel.Space, "%s: switch #%s panel space", source, sw.Id)
			if string(panel.Id) == switchPanels["group4"] {
				assert.Equal(t, "group4", panel.Name, "%s: panel name", source)
				assert.Equal(t, spaces["space4"], panel.Space.Id, "%s: panel space", source)
				assert.Equal(t, &panelPosition, panel.Position, "%s: panel position", source)
			} else {
				assert.NotEmpty(t, panel.Name, "%s: switch #%s panel name", source, sw.Id)
				assert.Nil(t, panel.Position, "%s: switch #%s panel position", source, sw.Id)
			}
		}

		chain := downstream(c, root)
		for _, sw := range chain.Switches {
			assertPanel(sw, "downstream")
		}
		for _, d := range chain.Devices {
			if string(d.Id) == devices["device8"] {
				assert.Equal(t, &devicePosition, d.Position, "downstream device8 position")
			} else {
				assert.Nil(t, d.Position, "downstream device #%s position", d.Id)
			}
		}

		for _, sw := range upstream(c, shared).Switches {
			assertPanel(sw, "upstream")
		}
	})

	spaceId := spaces["space2"]
	spacePath := fmt.Sprintf("/spaces/%s", spaceId)
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
			space string
			field string
			data  []byte
			want  int
		}{
			{spaceId, "image", []byte("plain text, not an image"), http.StatusUnsupportedMediaType},
			{spaceId, "file", pngImage(t, 2), http.StatusBadRequest},
			{spaceId, "image", make([]byte, 11<<20), http.StatusRequestEntityTooLarge},
			{uuid.Nil.String(), "image", pngImage(t, 2), http.StatusNotFound},
		} {
			req := backgroundImageRequest(t, tc.space, tc.field, tc.data)
			status := c.Send(req, nil)
			assert.Equal(t, tc.want, status, "upload %q (%d bytes) to space %s", tc.field, len(tc.data), tc.space)
		}
		status, _, _ := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusNotFound, status, "GET background image after rejected uploads")
	})

	t.Run("DeleteSpaceRemovesCrossSpaceReferences", func(t *testing.T) {
		c := c.with(t)
		removedPath := fmt.Sprintf("/spaces/%s", spaces["space5"])
		req := newRequest(t, http.MethodDelete, removedPath, nil)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "delete space5")

		for _, path := range []string{
			removedPath,
			fmt.Sprintf("/switch-panels/%s", switchPanels["group5"]),
			fmt.Sprintf("/switches/%s", switches["switch8"]),
			fmt.Sprintf("/switches/%s", right),
			fmt.Sprintf("/devices/%s", devices["device10"]),
			fmt.Sprintf("/devices/%s", devices["device11"]),
		} {
			req := newRequest(t, http.MethodGet, path, nil)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusNotFound, status, "GET %s after deleting space5", path)
		}

		want := idsOf(switches, "switch10", "switch7", "switch3", "switch1")
		assert.ElementsMatch(t, want, chainSwitchIds(upstream(c, shared).Switches), "upstream of device9 after deleting space5")

		remaining := slices.DeleteFunc(slices.Clone(allSwitches), func(name string) bool {
			return name == "switch8" || name == "switch11"
		})
		chain := downstream(c, root)
		logTree(t, root, chain)

		assert.ElementsMatch(t, idsOf(switches, remaining...), chainSwitchIds(chain.Switches), "downstream switches after deleting space5")
		assert.True(t, hasDeviceOnSwitch(chain, shared, left), "device9 should remain under switch10")
		assert.True(t, hasDeviceOnSwitch(chain, devices["device3"], switches["switch5"]), "device3 should be unaffected")
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
			ids    map[string]string
		}{
			{"/spaces", spaces},
			{"/switch-panels", switchPanels},
			{"/switches", switches},
			{"/devices", devices},
		} {
			for _, id := range tc.ids {
				path := fmt.Sprintf("%s/%s", tc.prefix, id)
				req := newRequest(t, http.MethodGet, path, nil)
				status := c.Send(req, nil)
				assert.Equal(t, http.StatusNotFound, status, "GET %s after deleting space group", path)
			}
		}
	})
}
