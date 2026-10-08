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

func downstream(c *client, breakerId int) ([]models.Breaker, []models.Device) {
	c.t.Helper()

	path := fmt.Sprintf("/breakers/%d/downstream", breakerId)
	var tree map[string]json.RawMessage
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &tree)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)

	var breakers []models.Breaker
	var devices []models.Device
	require.NoError(c.t, json.Unmarshal(tree["breakers"], &breakers), "decode downstream breakers")
	require.NoError(c.t, json.Unmarshal(tree["devices"], &devices), "decode downstream devices")
	return breakers, devices
}

func upstream(c *client, deviceId int) []models.Breaker {
	c.t.Helper()

	path := fmt.Sprintf("/devices/%d/upstream", deviceId)
	var breakers []models.Breaker
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &breakers)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)
	return breakers
}

func idsOf(ids map[string]int, names ...string) []int {
	result := make([]int, len(names))
	for i, name := range names {
		result[i] = ids[name]
	}
	return result
}

func breakerIds(breakers []models.Breaker) []int {
	result := make([]int, len(breakers))
	for i, b := range breakers {
		result[i] = b.Id
	}
	return result
}

func hasDeviceOnBreaker(devices []models.Device, deviceId, breakerId int) bool {
	return slices.ContainsFunc(devices, func(d models.Device) bool {
		return d.Id == deviceId && d.Breaker != nil && d.Breaker.Id == breakerId
	})
}

func logTree(t *testing.T, rootId int, breakers []models.Breaker, devices []models.Device) {
	t.Helper()

	children := make(map[int][]models.Breaker)
	names := make(map[int]string, len(breakers))
	for _, b := range breakers {
		names[b.Id] = b.Name
		if b.UpstreamBreaker != nil {
			children[b.UpstreamBreaker.Id] = append(children[b.UpstreamBreaker.Id], b)
		}
	}
	attached := make(map[int][]models.Device)
	for _, d := range devices {
		if d.Breaker != nil {
			attached[d.Breaker.Id] = append(attached[d.Breaker.Id], d)
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

	breakerGroups := make(map[string]int)
	for _, tc := range []struct{ space, name string }{
		{"space1", "group1"},
		{"space2", "group2"},
		{"space3", "group3"},
		{"space4", "group4"},
		{"space5", "group5"},
	} {
		path := fmt.Sprintf("/spaces/%d/breaker-groups", spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var breakerGroup models.BreakerGroup
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &breakerGroup)
		require.Equal(t, http.StatusOK, status, "create breaker group %s", tc.name)
		breakerGroups[tc.name] = breakerGroup.Id
	}

	breakers := make(map[string]int)
	for _, tc := range []struct{ group, name, upstream string }{
		{"group1", "breaker1", ""},
		{"group1", "breaker2", "breaker1"},
		{"group1", "breaker3", "breaker1"},
		{"group2", "breaker4", "breaker2"},
		{"group2", "breaker5", "breaker2"},
		{"group3", "breaker6", "breaker2"},
		{"group4", "breaker7", "breaker3"},
		{"group5", "breaker8", "breaker3"},
		{"group2", "breaker9", "breaker5"},
		{"group4", "breaker10", "breaker7"},
		{"group5", "breaker11", "breaker8"},
	} {
		path := fmt.Sprintf("/breaker-groups/%d/breakers", breakerGroups[tc.group])
		body := map[string]any{"name": tc.name}
		if tc.upstream != "" {
			body["upstream_breaker_id"] = breakers[tc.upstream]
		}

		var breaker models.Breaker
		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, &breaker)
		require.Equal(t, http.StatusOK, status, "create breaker %s", tc.name)
		breakers[tc.name] = breaker.Id
	}

	root := breakers["breaker1"]
	left := breakers["breaker10"]
	right := breakers["breaker11"]

	t.Run("BreakerUpstreamRejectsCycle", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct{ breaker, upstream string }{
			{"breaker1", "breaker10"},
			{"breaker10", "breaker10"},
		} {
			path := fmt.Sprintf("/breakers/%d/upstream", breakers[tc.breaker])
			body := map[string]any{"upstream_breaker_id": breakers[tc.upstream]}

			req := newRequest(t, http.MethodPatch, path, body)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusBadRequest, status, "set upstream of %s to %s", tc.breaker, tc.upstream)
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

	for _, tc := range []struct{ device, breaker string }{
		{"device1", "breaker4"},
		{"device2", "breaker4"},
		{"device3", "breaker5"},
		{"device4", "breaker9"},
		{"device5", "breaker6"},
		{"device6", "breaker6"},
		{"device7", "breaker6"},
		{"device8", "breaker7"},
		{"device9", "breaker10"},
		{"device9", "breaker11"},
		{"device10", "breaker8"},
		{"device11", "breaker8"},
	} {
		path := fmt.Sprintf("/devices/%d/breakers", devices[tc.device])
		body := map[string]any{"breaker_id": breakers[tc.breaker]}

		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "attach %s to %s", tc.device, tc.breaker)
	}

	t.Run("AttachDeviceRejectsDuplicate", func(t *testing.T) {
		c := c.with(t)
		path := fmt.Sprintf("/devices/%d/breakers", shared)
		body := map[string]any{"breaker_id": left}

		req := newRequest(t, http.MethodPost, path, body)
		status := c.Send(req, nil)
		assert.Equal(t, http.StatusConflict, status, "attach device9 to breaker10 again")
	})

	allBreakers := make([]string, 0, len(breakers))
	for name := range breakers {
		allBreakers = append(allBreakers, name)
	}

	t.Run("BreakerDownstreamIncludesDescendants", func(t *testing.T) {
		c := c.with(t)
		treeBreakers, treeDevices := downstream(c, root)
		logTree(t, root, treeBreakers, treeDevices)

		assert.ElementsMatch(t, idsOf(breakers, allBreakers...), breakerIds(treeBreakers), "downstream breakers")
		assert.True(t, hasDeviceOnBreaker(treeDevices, shared, left), "device9 should appear under breaker10")
		assert.True(t, hasDeviceOnBreaker(treeDevices, shared, right), "device9 should appear under breaker11")
	})

	t.Run("DeviceUpstreamDeduplicatesSharedPath", func(t *testing.T) {
		c := c.with(t)
		want := idsOf(breakers, "breaker10", "breaker7", "breaker11", "breaker8", "breaker3", "breaker1")
		assert.ElementsMatch(t, want, breakerIds(upstream(c, shared)), "upstream of device9")
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
			fmt.Sprintf("/breaker-groups/%d", breakerGroups["group5"]),
			fmt.Sprintf("/breakers/%d", breakers["breaker8"]),
			fmt.Sprintf("/breakers/%d", right),
			fmt.Sprintf("/devices/%d", devices["device10"]),
			fmt.Sprintf("/devices/%d", devices["device11"]),
		} {
			req := newRequest(t, http.MethodGet, path, nil)
			status := c.Send(req, nil)
			assert.Equal(t, http.StatusNotFound, status, "GET %s after deleting space5", path)
		}

		want := idsOf(breakers, "breaker10", "breaker7", "breaker3", "breaker1")
		assert.ElementsMatch(t, want, breakerIds(upstream(c, shared)), "upstream of device9 after deleting space5")

		remaining := slices.DeleteFunc(slices.Clone(allBreakers), func(name string) bool {
			return name == "breaker8" || name == "breaker11"
		})
		treeBreakers, treeDevices := downstream(c, root)
		logTree(t, root, treeBreakers, treeDevices)

		assert.ElementsMatch(t, idsOf(breakers, remaining...), breakerIds(treeBreakers), "downstream breakers after deleting space5")
		assert.True(t, hasDeviceOnBreaker(treeDevices, shared, left), "device9 should remain under breaker10")
		assert.True(t, hasDeviceOnBreaker(treeDevices, devices["device3"], breakers["breaker5"]), "device3 should be unaffected")
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
			{"/breaker-groups", breakerGroups},
			{"/breakers", breakers},
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
