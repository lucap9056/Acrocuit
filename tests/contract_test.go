package test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

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
	spaceGroupPath := fmt.Sprintf("/space-groups/%s", spaceGroupId)
	t.Cleanup(func() {
		req := newRequest(t, http.MethodDelete, spaceGroupPath, nil)
		c.Send(req, nil)
	})

	body = map[string]any{"name": "space"}
	req = newRequest(t, http.MethodPost, spaceGroupPath+"/spaces", body)
	data = c.checkShape(req, http.StatusOK, successShape(spaceShape()))
	spaceId := mustId(t, data)
	expectText(t, data, spaceGroupId, "space_group", "id")
	spacePath := fmt.Sprintf("/spaces/%s", spaceId)

	body = map[string]any{"name": "switch panel", "position": map[string]int{"x": 1, "y": 2, "z": 3}, "attributes": map[string]any{"location": "B1"}}
	req = newRequest(t, http.MethodPost, spacePath+"/switch-panels", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
	switchPanelId := mustId(t, data)
	expectText(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	expectText(t, data, "B1", "attributes", "location")
	switchPanelPath := fmt.Sprintf("/switch-panels/%s", switchPanelId)

	body = map[string]any{"name": "root", "attributes": map[string]any{"rating": 32}}
	req = newRequest(t, http.MethodPost, switchPanelPath+"/switches", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
	rootId := mustId(t, data)
	expectText(t, data, switchPanelId, "switch_panel", "id")
	expectInt(t, data, 32, "attributes", "rating")
	rootPath := fmt.Sprintf("/switches/%s", rootId)

	body = map[string]any{"name": "child"}
	req = newRequest(t, http.MethodPost, switchPanelPath+"/switches", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
	childId := mustId(t, data)
	expectEmptyObject(t, data, "attributes")
	childPath := fmt.Sprintf("/switches/%s", childId)

	body = map[string]any{"upstream_switch_id": rootId}
	req = newRequest(t, http.MethodPost, childPath+"/upstreams", body)
	data = c.checkShape(req, http.StatusOK, successShape(switchLinkShape()))
	expectText(t, data, rootId, "upstream_switch_id")
	expectText(t, data, childId, "downstream_switch_id")
	upstreamPath := fmt.Sprintf("%s/upstreams/%s", childPath, rootId)

	body = map[string]any{"name": "device", "position": map[string]int{"x": 1, "y": 2, "z": 3}}
	req = newRequest(t, http.MethodPost, spacePath+"/devices", body)
	data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
	deviceId := mustId(t, data)
	expectText(t, data, spaceId, "space", "id")
	expectPosition(t, data, 1, 2, 3)
	devicePath := fmt.Sprintf("/devices/%s", deviceId)

	body = map[string]any{"switch_id": childId, "attributes": map[string]any{"socket": "A1"}}
	req = newRequest(t, http.MethodPost, devicePath+"/switches", body)
	data = c.checkShape(req, http.StatusOK, successShape(deviceLinkShape()))
	expectText(t, data, "A1", "attributes", "socket")

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
		expectText(t, data, spaceGroupId, "space_group", "id")

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
		expectText(t, data, spaceId, "space", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/switch-panels", nil, arrayOf(switchPanelShape())},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		req = newRequest(t, http.MethodPut, switchPanelPath, map[string]any{"name": "renamed"})
		data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
		expectText(t, data, "renamed", "name")
		expectText(t, data, "B1", "attributes", "location")

		body := map[string]any{"attributes": map[string]any{"location": "B2"}}
		req = newRequest(t, http.MethodPut, switchPanelPath, body)
		data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
		expectText(t, data, "renamed", "name")
		expectText(t, data, "B2", "attributes", "location")

		body = map[string]any{"position": map[string]int{"x": 4, "y": 5, "z": 6}}
		req = newRequest(t, http.MethodPut, switchPanelPath+"/position", body)
		data = c.checkShape(req, http.StatusOK, successShape(switchPanelShape()))
		expectPosition(t, data, 4, 5, 6)
		expectText(t, data, "B2", "attributes", "location")
	})

	t.Run("Switches", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, childPath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectText(t, data, switchPanelId, "switch_panel", "id")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, switchPanelPath + "/switches", nil, arrayOf(switchShape())},
			{http.MethodPatch, childPath + "/order", map[string]any{"display_order": 1}, arrayOf(switchShape())},
			{http.MethodGet, childPath + "/upstreams", nil, arrayOf(linkedSwitchShape())},
			{http.MethodDelete, upstreamPath, nil, equals(nil)},
			{http.MethodPost, childPath + "/upstreams", map[string]any{"upstream_switch_id": rootId}, switchLinkShape()},
		} {
			req := newRequest(t, tc.method, tc.path, tc.body)
			c.checkShape(req, http.StatusOK, successShape(tc.data))
		}

		body := map[string]any{"name": "renamed", "attributes": map[string]any{"rating": 16}}
		req = newRequest(t, http.MethodPut, childPath, body)
		data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectText(t, data, "renamed", "name")
		expectInt(t, data, 16, "attributes", "rating")

		req = newRequest(t, http.MethodPut, childPath, map[string]any{"name": "renamed again"})
		data = c.checkShape(req, http.StatusOK, successShape(switchShape()))
		expectInt(t, data, 16, "attributes", "rating")

		body = map[string]any{"attributes": map[string]any{"cable": "5.5mm2"}}
		req = newRequest(t, http.MethodPut, upstreamPath, body)
		data = c.checkShape(req, http.StatusOK, successShape(switchLinkShape()))
		expectText(t, data, "5.5mm2", "attributes", "cable")

		req = newRequest(t, http.MethodGet, childPath+"/upstreams", nil)
		data = c.checkShape(req, http.StatusOK, successShape(arrayOf(linkedSwitchShape())))
		expectText(t, data, rootId, "0", "id")
		expectInt(t, data, 32, "0", "attributes", "rating")
		expectText(t, data, "5.5mm2", "0", "link", "attributes", "cable")

		req = newRequest(t, http.MethodGet, rootPath+"/downstream", nil)
		data = c.checkShape(req, http.StatusOK, successShape(powerChainShape()))
		expectText(t, data, rootId, "switch_links", "0", "upstream_switch_id")
		expectText(t, data, "5.5mm2", "switch_links", "0", "attributes", "cable")
		expectText(t, data, childId, "device_links", "0", "switch_id")
		expectText(t, data, "A1", "device_links", "0", "attributes", "socket")
	})

	t.Run("Devices", func(t *testing.T) {
		c := c.with(t)
		req := newRequest(t, http.MethodGet, devicePath, nil)
		data := c.checkShape(req, http.StatusOK, successShape(deviceShape()))
		expectText(t, data, spaceId, "space", "id")

		body := map[string]any{"position": map[string]int{"x": 4, "y": 5, "z": 6}}
		req = newRequest(t, http.MethodPut, devicePath+"/position", body)
		data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
		expectPosition(t, data, 4, 5, 6)

		body = map[string]any{"name": "renamed", "attributes": map[string]any{"watts": 1500}}
		req = newRequest(t, http.MethodPut, devicePath, body)
		data = c.checkShape(req, http.StatusOK, successShape(deviceShape()))
		expectInt(t, data, 1500, "attributes", "watts")
		expectPosition(t, data, 4, 5, 6)

		detachPath := fmt.Sprintf("%s/switches/%s", devicePath, childId)
		body = map[string]any{"attributes": map[string]any{"socket": "B2"}}
		req = newRequest(t, http.MethodPut, detachPath, body)
		data = c.checkShape(req, http.StatusOK, successShape(deviceLinkShape()))
		expectText(t, data, "B2", "attributes", "socket")

		req = newRequest(t, http.MethodGet, devicePath+"/switches", nil)
		data = c.checkShape(req, http.StatusOK, successShape(arrayOf(linkedSwitchShape())))
		expectText(t, data, childId, "0", "id")
		expectInt(t, data, 16, "0", "attributes", "rating")
		expectText(t, data, "B2", "0", "link", "attributes", "socket")

		req = newRequest(t, http.MethodGet, devicePath+"/upstream", nil)
		data = c.checkShape(req, http.StatusOK, successShape(powerChainShape()))
		expectText(t, data, deviceId, "devices", "0", "id")
		expectInt(t, data, 1500, "devices", "0", "attributes", "watts")
		expectText(t, data, "B2", "device_links", "0", "attributes", "socket")
		expectText(t, data, "5.5mm2", "switch_links", "0", "attributes", "cable")

		for _, tc := range []struct {
			method, path string
			body         any
			data         shape
		}{
			{http.MethodGet, spacePath + "/devices", nil, arrayOf(deviceShape())},
			{http.MethodGet, devicePath + "/switches", nil, arrayOf(linkedSwitchShape())},
			{http.MethodDelete, detachPath, nil, equals(nil)},
			{http.MethodPost, devicePath + "/switches", map[string]any{"switch_id": childId}, deviceLinkShape()},
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
			{http.MethodGet, "/spaces/" + uuid.Nil.String(), nil, http.StatusNotFound},
			{http.MethodGet, "/spaces/abc", nil, http.StatusBadRequest},
			{http.MethodPost, "/space-groups", map[string]any{}, http.StatusBadRequest},
			{http.MethodPost, rootPath + "/upstreams", map[string]any{"upstream_switch_id": rootId}, http.StatusBadRequest},
			{http.MethodPost, rootPath + "/upstreams", map[string]any{"upstream_switch_id": childId}, http.StatusBadRequest},
			{http.MethodPost, childPath + "/upstreams", map[string]any{"upstream_switch_id": rootId}, http.StatusConflict},
			{http.MethodPut, rootPath + "/upstreams/" + uuid.Nil.String(), map[string]any{"attributes": map[string]any{}}, http.StatusNotFound},
			{http.MethodPut, upstreamPath, map[string]any{}, http.StatusBadRequest},
			{http.MethodPut, devicePath, map[string]any{"attributes": []int{1}}, http.StatusBadRequest},
			{http.MethodPut, devicePath, map[string]any{"attributes": "text"}, http.StatusBadRequest},
			{http.MethodPut, devicePath, map[string]any{"attributes": map[string]string{"note": strings.Repeat("x", 2048)}}, http.StatusBadRequest},
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
