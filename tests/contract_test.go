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

func breakerGroupShape() shape {
	return object(map[string]shape{
		"id":       integer,
		"name":     text,
		"space":    nullable(spaceShape()),
		"position": nullable(positionShape()),
	})
}

func breakerShape() shape {
	return object(map[string]shape{
		"id":               integer,
		"name":             text,
		"display_order":    integer,
		"breaker_group":    nullable(breakerGroupShape()),
		"upstream_breaker": nullable(lazy(breakerShape)),
	})
}

func deviceShape() shape {
	return object(map[string]shape{
		"id":       integer,
		"name":     text,
		"space":    nullable(spaceShape()),
		"breaker":  nullable(breakerShape()),
		"position": nullable(positionShape()),
	})
}

func downstreamShape() shape {
	return object(map[string]shape{
		"breakers": arrayOf(breakerShape()),
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

	body = map[string]any{"name": "breaker group", "position": map[string]int{"x": 1, "y": 2, "z": 3}}
	req = newRequest(t, http.MethodPost, spacePath+"/breaker-groups", body)
	data = c.checkShape(req, http.StatusOK, successShape(breakerGroupShape()))
	breakerGroupId := mustId(t, data)
	expectInt(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	breakerGroupPath := fmt.Sprintf("/breaker-groups/%d", breakerGroupId)

	body = map[string]any{"name": "root"}
	req = newRequest(t, http.MethodPost, breakerGroupPath+"/breakers", body)
	data = c.checkShape(req, http.StatusOK, successShape(breakerShape()))
	rootId := mustId(t, data)
	expectInt(t, data, breakerGroupId, "breaker_group", "id")
	expectNull(t, data, "upstream_breaker")
	rootPath := fmt.Sprintf("/breakers/%d", rootId)

	body = map[string]any{"name": "child", "upstream_breaker_id": rootId}
	req = newRequest(t, http.MethodPost, breakerGroupPath+"/breakers", body)
	data = c.checkShape(req, http.StatusOK, successShape(breakerShape()))
	childId := mustId(t, data)
	expectInt(t, data, rootId, "upstream_breaker", "id")
	childPath := fmt.Sprintf("/breakers/%d", childId)

	body = map[string]any{"name": "device", "position": map[string]int{"x": 1, "y": 2, "z": 3}}
	req = newRequest(t, http.MethodPost, spacePath+"/devices", body)
	data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
	deviceId := mustId(t, data)
	expectInt(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	devicePath := fmt.Sprintf("/devices/%d", deviceId)

	body = map[string]any{"breaker_id": childId}
	req = newRequest(t, http.MethodPost, devicePath+"/breakers", body)
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
			{http.MethodPatch, spacePath + "/order", map[string]any{"display_order": 1}, spaceShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		req = backgroundImageRequest(t, spaceId, "image", pngImage(t, 2))
		c.checkShape(req, http.StatusOK, successShape(spaceShape()))
		req = newRequest(t, http.MethodDelete, spacePath+"/background-image", nil)
		c.checkShape(req, http.StatusOK, successShape(spaceShape()))
	})

	t.Run("BreakerGroups", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, breakerGroupPath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(breakerGroupShape()))
		expectInt(t, data, spaceId, "space", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/breaker-groups", nil, arrayOf(breakerGroupShape())},
			{http.MethodPut, breakerGroupPath, map[string]any{"name": "renamed"}, breakerGroupShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		body := map[string]any{"position": map[string]int{"x": 4, "y": 5, "z": 6}}
		req = newRequest(t, http.MethodPut, breakerGroupPath+"/position", body)
		data = c.checkShape(req, http.StatusOK, successShape(breakerGroupShape()))
		expectPosition(t, data, 4, 5, 6)
	})

	t.Run("Breakers", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, childPath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(breakerShape()))
		expectInt(t, data, breakerGroupId, "breaker_group", "id")
		expectInt(t, data, rootId, "upstream_breaker", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, breakerGroupPath + "/breakers", nil, arrayOf(breakerShape())},
			{http.MethodPut, childPath, map[string]any{"name": "renamed"}, breakerShape()},
			{http.MethodPatch, childPath + "/order", map[string]any{"display_order": 1}, breakerShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		body := map[string]any{"upstream_breaker_id": nil}
		req = newRequest(t, http.MethodPatch, childPath+"/upstream", body)
		data = c.checkShape(req, http.StatusOK, successShape(breakerShape()))
		expectNull(t, data, "upstream_breaker")

		body = map[string]any{"upstream_breaker_id": rootId}
		req = newRequest(t, http.MethodPatch, childPath+"/upstream", body)
		data = c.checkShape(req, http.StatusOK, successShape(breakerShape()))
		expectInt(t, data, rootId, "upstream_breaker", "id")

		req = newRequest(t, http.MethodGet, rootPath+"/downstream", nil)
		data = c.checkShape(req, http.StatusOK, successShape(downstreamShape()))
		expectInt(t, data, childId, "devices", "0", "breaker", "id")
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

		detachPath := fmt.Sprintf("%s/breakers/%d", devicePath, childId)
		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/devices", nil, arrayOf(deviceShape())},
			{http.MethodPut, devicePath, map[string]any{"name": "renamed"}, deviceShape()},
			{http.MethodGet, devicePath + "/breakers", nil, arrayOf(breakerShape())},
			{http.MethodGet, devicePath + "/upstream", nil, arrayOf(breakerShape())},
			{http.MethodDelete, detachPath, nil, equals(nil)},
			{http.MethodPost, devicePath + "/breakers", map[string]any{"breaker_id": childId}, equals(nil)},
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
			{http.MethodPatch, rootPath + "/upstream", map[string]any{"upstream_breaker_id": rootId}, http.StatusBadRequest},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, tc.want, failureShape())
		}
	})

	t.Run("Deletes", func(t *testing.T) {
		c := c.with(t)
		for _, path := range []string{devicePath, childPath, breakerGroupPath, spacePath, spaceGroupPath} {
			req := newRequest(t, http.MethodDelete, path, nil)
			c.checkShape(req, http.StatusOK, successShape(equals(nil)))
		}
	})
}
