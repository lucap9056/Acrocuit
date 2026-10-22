package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shape func(t *testing.T, path string, v any)

func integer(t *testing.T, path string, v any) {
	t.Helper()
	if !assert.IsType(t, json.Number(""), v, path) {
		return
	}
	_, err := v.(json.Number).Int64()
	assert.NoError(t, err, "%s: want integer", path)
}

func boolean(t *testing.T, path string, v any) {
	t.Helper()
	assert.IsType(t, true, v, path)
}

func text(t *testing.T, path string, v any) {
	t.Helper()
	assert.IsType(t, "", v, path)
}

func equals(want any) shape {
	return func(t *testing.T, path string, v any) {
		t.Helper()
		assert.Equal(t, want, v, path)
	}
}

func object(fields map[string]shape) shape {
	return func(t *testing.T, path string, v any) {
		t.Helper()
		if !assert.IsType(t, map[string]any{}, v, path) {
			return
		}
		obj := v.(map[string]any)
		assert.ElementsMatch(t, slices.Collect(maps.Keys(fields)), slices.Collect(maps.Keys(obj)), "%s: fields", path)
		for key, check := range fields {
			if value, ok := obj[key]; ok {
				check(t, path+"."+key, value)
			}
		}
	}
}

func nullable(check shape) shape {
	return func(t *testing.T, path string, v any) {
		t.Helper()
		if v != nil {
			check(t, path, v)
		}
	}
}

func arrayOf(check shape) shape {
	return func(t *testing.T, path string, v any) {
		t.Helper()
		if !assert.IsType(t, []any{}, v, path) {
			return
		}
		for i, item := range v.([]any) {
			check(t, fmt.Sprintf("%s[%d]", path, i), item)
		}
	}
}

func lazy(build func() shape) shape {
	return func(t *testing.T, path string, v any) {
		t.Helper()
		build()(t, path, v)
	}
}

func spaceGroupShape() shape {
	return object(map[string]shape{
		"id":   integer,
		"name": text,
	})
}

func spaceShape() shape {
	return object(map[string]shape{
		"id":                          integer,
		"name":                        text,
		"display_order":               integer,
		"background_image_updated_at": integer,
		"space_group":                 nullable(spaceGroupShape()),
	})
}

func positionShape() shape {
	return object(map[string]shape{
		"x": integer,
		"y": integer,
		"z": integer,
	})
}

func switchPanelShape() shape {
	return object(map[string]shape{
		"id":       integer,
		"name":     text,
		"space":    nullable(spaceShape()),
		"position": nullable(positionShape()),
	})
}

func switchShape() shape {
	return object(map[string]shape{
		"id":              integer,
		"name":            text,
		"display_order":   integer,
		"switch_panel":    nullable(switchPanelShape()),
		"upstream_switch": nullable(lazy(switchShape)),
	})
}

func deviceShape() shape {
	return object(map[string]shape{
		"id":       integer,
		"name":     text,
		"space":    nullable(spaceShape()),
		"switch":   nullable(switchShape()),
		"position": nullable(positionShape()),
	})
}

func metaShape() shape {
	return object(map[string]shape{
		"auth_required": boolean,
	})
}

func downstreamShape() shape {
	return object(map[string]shape{
		"switches": arrayOf(switchShape()),
		"devices":  arrayOf(deviceShape()),
	})
}

func successShape(data shape) shape {
	return object(map[string]shape{
		"success": equals(true),
		"data":    data,
	})
}

func failureShape() shape {
	return object(map[string]shape{
		"success": equals(false),
		"data":    equals(nil),
		"error":   text,
	})
}

func (c *client) checkShape(req *http.Request, wantStatus int, envelope shape) any {
	c.t.Helper()

	status, _, raw := c.do(req)
	require.Equal(c.t, wantStatus, status, "%s %s", req.Method, req.URL.Path)

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var body map[string]any
	require.NoError(c.t, decoder.Decode(&body), "%s %s: decode body (%s)", req.Method, req.URL.Path, raw)
	envelope(c.t, req.Method+" "+req.URL.Path, body)
	return body["data"]
}

func field(t *testing.T, v any, keys ...string) any {
	t.Helper()
	for i, key := range keys {
		path := "data." + strings.Join(keys[:i], ".")
		switch node := v.(type) {
		case map[string]any:
			v = node[key]
		case []any:
			index, err := strconv.Atoi(key)
			if !assert.True(t, err == nil && index >= 0 && index < len(node), "%s: index %q out of range (len %d)", path, key, len(node)) {
				return nil
			}
			v = node[index]
		default:
			assert.Fail(t, fmt.Sprintf("%s: got %T, want object or array", path, v))
			return nil
		}
	}
	return v
}

func intField(t *testing.T, v any, keys ...string) (int, bool) {
	t.Helper()
	path := "data." + strings.Join(keys, ".")
	value := field(t, v, keys...)
	if !assert.IsType(t, json.Number(""), value, path) {
		return 0, false
	}
	n, err := value.(json.Number).Int64()
	if !assert.NoError(t, err, "%s: want integer", path) {
		return 0, false
	}
	return int(n), true
}

func expectInt(t *testing.T, v any, want int, keys ...string) {
	t.Helper()
	if got, ok := intField(t, v, keys...); ok {
		assert.Equal(t, want, got, "data.%s", strings.Join(keys, "."))
	}
}

func expectNull(t *testing.T, v any, keys ...string) {
	t.Helper()
	assert.Nil(t, field(t, v, keys...), "data.%s", strings.Join(keys, "."))
}

func expectPosition(t *testing.T, v any, x, y, z int) {
	t.Helper()
	expectInt(t, v, x, "position", "x")
	expectInt(t, v, y, "position", "y")
	expectInt(t, v, z, "position", "z")
}

func mustId(t *testing.T, v any) int {
	t.Helper()
	id, ok := intField(t, v, "id")
	require.True(t, ok, "data.id")
	return id
}

func TestMetaContract(t *testing.T) {
	c := newClient(t)
	c.identityToken = ""
	c.accessToken = ""

	req := newRequest(t, http.MethodGet, "/meta", nil)
	c.checkShape(req, http.StatusOK, successShape(metaShape()))
}

func TestContract(t *testing.T) {
	c := newClient(t)

	body := map[string]any{"name": t.Name()}
	req := newRequest(t, http.MethodPost, "/space-groups", body)
	data := c.checkShape(req, http.StatusOK, successShape(spaceGroupShape()))
	spaceGroupId := mustId(t, data)
	spaceGroupPath := fmt.Sprintf("/space-groups/%d", spaceGroupId)
	t.Cleanup(func() {
		req := newRequest(t, http.MethodDelete, spaceGroupPath, nil)
		c.Send(req, nil)
	})

	body = map[string]any{"name": "space"}
	req = newRequest(t, http.MethodPost, spaceGroupPath+"/spaces", body)
	data = c.checkShape(req, http.StatusOK, successShape(spaceShape()))
	spaceId := mustId(t, data)
	expectInt(t, data, spaceGroupId, "space_group", "id")
	spacePath := fmt.Sprintf("/spaces/%d", spaceId)

	body = map[string]any{"name": "switch panel", "position": map[string]int{"x": 1, "y": 2, "z": 3}}
	req = newRequest(t, http.MethodPost, spacePath+"/switch-panels", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
	switchPanelId := mustId(t, data)
	expectInt(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	switchPanelPath := fmt.Sprintf("/switch-panels/%d", switchPanelId)

	body = map[string]any{"name": "root"}
	req = newRequest(t, http.MethodPost, switchPanelPath+"/switches", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
	rootId := mustId(t, data)
	expectInt(t, data, switchPanelId, "switch_panel", "id")
	expectNull(t, data, "upstream_switch")
	rootPath := fmt.Sprintf("/switches/%d", rootId)

	body = map[string]any{"name": "child", "upstream_switch_id": rootId}
	req = newRequest(t, http.MethodPost, switchPanelPath+"/switches", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
	childId := mustId(t, data)
	expectInt(t, data, rootId, "upstream_switch", "id")
	childPath := fmt.Sprintf("/switches/%d", childId)

	body = map[string]any{"name": "device", "position": map[string]int{"x": 1, "y": 2, "z": 3}}
	req = newRequest(t, http.MethodPost, spacePath+"/devices", body)
	data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
	deviceId := mustId(t, data)
	expectInt(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	devicePath := fmt.Sprintf("/devices/%d", deviceId)

	body = map[string]any{"switch_id": childId}
	req = newRequest(t, http.MethodPost, devicePath+"/switches", body)
	c.checkShape(req, http.StatusOK, successShape(equals(nil)))

	t.Run("SpaceGroups", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, "/space-groups", nil, arrayOf(spaceGroupShape())},
			{http.MethodGet, spaceGroupPath, nil, spaceGroupShape()},
			{http.MethodPut, spaceGroupPath, map[string]any{"name": "renamed"}, spaceGroupShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}
	})

	t.Run("Spaces", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, spacePath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(spaceShape()))
		expectInt(t, data, spaceGroupId, "space_group", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spaceGroupPath + "/spaces", nil, arrayOf(spaceShape())},
			{http.MethodPut, spacePath, map[string]any{"name": "renamed"}, spaceShape()},
			{http.MethodPatch, spacePath + "/order", map[string]any{"display_order": 1}, arrayOf(spaceShape())},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		req = backgroundImageRequest(t, spaceId, "image", pngImage(t, 2))
		c.checkShape(req, http.StatusOK, successShape(spaceShape()))
		req = newRequest(t, http.MethodDelete, spacePath+"/background-image", nil)
		c.checkShape(req, http.StatusOK, successShape(spaceShape()))
	})

	t.Run("SwitchPanels", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, switchPanelPath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
		expectInt(t, data, spaceId, "space", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/switch-panels", nil, arrayOf(switchPanelShape())},
			{http.MethodPut, switchPanelPath, map[string]any{"name": "renamed"}, switchPanelShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		body := map[string]any{"position": map[string]int{"x": 4, "y": 5, "z": 6}}
		req = newRequest(t, http.MethodPut, switchPanelPath+"/position", body)
		data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
		expectPosition(t, data, 4, 5, 6)
	})

	t.Run("Switches", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, childPath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectInt(t, data, switchPanelId, "switch_panel", "id")
		expectInt(t, data, rootId, "upstream_switch", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, switchPanelPath + "/switches", nil, arrayOf(switchShape())},
			{http.MethodPut, childPath, map[string]any{"name": "renamed"}, switchShape()},
			{http.MethodPatch, childPath + "/order", map[string]any{"display_order": 1}, arrayOf(switchShape())},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		body := map[string]any{"upstream_switch_id": nil}
		req = newRequest(t, http.MethodPatch, childPath+"/upstream", body)
		data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectNull(t, data, "upstream_switch")

		body = map[string]any{"upstream_switch_id": rootId}
		req = newRequest(t, http.MethodPatch, childPath+"/upstream", body)
		data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectInt(t, data, rootId, "upstream_switch", "id")

		req = newRequest(t, http.MethodGet, rootPath+"/downstream", nil)
		data = c.checkShape(req, http.StatusOK, successShape(downstreamShape()))
		expectInt(t, data, childId, "devices", "0", "switch", "id")
	})

	t.Run("Devices", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, devicePath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(deviceShape()))
		expectInt(t, data, spaceId, "space", "id")

		body := map[string]any{"position": map[string]int{"x": 4, "y": 5, "z": 6}}
		req = newRequest(t, http.MethodPut, devicePath+"/position", body)
		data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
		expectPosition(t, data, 4, 5, 6)

		detachPath := fmt.Sprintf("%s/switches/%d", devicePath, childId)
		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/devices", nil, arrayOf(deviceShape())},
			{http.MethodPut, devicePath, map[string]any{"name": "renamed"}, deviceShape()},
			{http.MethodGet, devicePath + "/switches", nil, arrayOf(switchShape())},
			{http.MethodGet, devicePath + "/upstream", nil, arrayOf(switchShape())},
			{http.MethodDelete, detachPath, nil, equals(nil)},
			{http.MethodPost, devicePath + "/switches", map[string]any{"switch_id": childId}, equals(nil)},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}
	})

	t.Run("Errors", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct {
			method, path string
			body         any
			want         int
		}{
			{http.MethodGet, "/spaces/0", nil, http.StatusNotFound},
			{http.MethodGet, "/spaces/abc", nil, http.StatusBadRequest},
			{http.MethodPost, "/space-groups", map[string]any{}, http.StatusBadRequest},
			{http.MethodPatch, rootPath + "/upstream", map[string]any{"upstream_switch_id": rootId}, http.StatusBadRequest},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, tc.want, failureShape())
		}
	})

	t.Run("Deletes", func(t *testing.T) {
		c := c.with(t)
		for _, path := range []string{devicePath, childPath, switchPanelPath, spacePath, spaceGroupPath} {
			req := newRequest(t, http.MethodDelete, path, nil)
			c.checkShape(req, http.StatusOK, successShape(equals(nil)))
		}
	})
}
