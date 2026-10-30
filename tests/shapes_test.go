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

func attributes(t *testing.T, path string, v any) {
	t.Helper()
	assert.IsType(t, map[string]any{}, v, path)
}

func spaceGroupShape() shape {
	return object(map[string]shape{
		"id":   text,
		"name": text,
		"role": text,
	})
}

func spaceShape() shape {
	return object(map[string]shape{
		"id":                          text,
		"name":                        text,
		"display_order":               integer,
		"background_image_updated_at": integer,
		"space_group":                 nullable(refShape()),
	})
}

func positionShape() shape {
	return object(map[string]shape{
		"x": integer,
		"y": integer,
		"z": integer,
	})
}

func refShape() shape {
	return object(map[string]shape{
		"id": text,
	})
}

func switchPanelShape() shape {
	return switchPanelShapeWith(attributes)
}

func chainSwitchPanelShape() shape {
	return switchPanelShapeWith(nullable(attributes))
}

func switchPanelShapeWith(attributesShape shape) shape {
	return object(map[string]shape{
		"id":         text,
		"name":       text,
		"attributes": attributesShape,
		"space":      nullable(refShape()),
		"position":   nullable(positionShape()),
	})
}

func switchShape() shape {
	return object(map[string]shape{
		"id":            text,
		"name":          text,
		"display_order": integer,
		"attributes":    attributes,
		"switch_panel":  nullable(refShape()),
	})
}

func chainSwitchShape() shape {
	return object(map[string]shape{
		"id":            text,
		"name":          text,
		"display_order": integer,
		"attributes":    attributes,
		"switch_panel":  nullable(chainSwitchPanelShape()),
	})
}

func linkedSwitchShape() shape {
	return object(map[string]shape{
		"id":            text,
		"name":          text,
		"display_order": integer,
		"attributes":    attributes,
		"switch_panel":  nullable(refShape()),
		"link": object(map[string]shape{
			"attributes": attributes,
		}),
	})
}

func deviceShape() shape {
	return object(map[string]shape{
		"id":         text,
		"name":       text,
		"attributes": attributes,
		"space":      nullable(refShape()),
		"position":   nullable(positionShape()),
	})
}

func switchLinkShape() shape {
	return object(map[string]shape{
		"upstream_switch_id":   text,
		"downstream_switch_id": text,
		"attributes":           attributes,
	})
}

func deviceLinkShape() shape {
	return object(map[string]shape{
		"device_id":  text,
		"switch_id":  text,
		"attributes": attributes,
	})
}

func metaShape() shape {
	return object(map[string]shape{
		"auth_required": boolean,
	})
}

func powerChainShape() shape {
	return object(map[string]shape{
		"switches":     arrayOf(chainSwitchShape()),
		"switch_links": arrayOf(switchLinkShape()),
		"devices":      arrayOf(deviceShape()),
		"device_links": arrayOf(deviceLinkShape()),
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

func expectText(t *testing.T, v any, want string, keys ...string) {
	t.Helper()
	assert.Equal(t, want, field(t, v, keys...), "data.%s", strings.Join(keys, "."))
}

func expectEmptyObject(t *testing.T, v any, keys ...string) {
	t.Helper()
	assert.Equal(t, map[string]any{}, field(t, v, keys...), "data.%s", strings.Join(keys, "."))
}

func expectPosition(t *testing.T, v any, x, y, z int) {
	t.Helper()
	expectInt(t, v, x, "position", "x")
	expectInt(t, v, y, "position", "y")
	expectInt(t, v, z, "position", "z")
}

func mustId(t *testing.T, v any) string {
	t.Helper()
	id, ok := field(t, v, "id").(string)
	require.True(t, ok, "data.id")
	return id
}
