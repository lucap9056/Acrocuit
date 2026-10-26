package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	host             = env("ACROCUIT_HOST", "http://localhost:8080")
	apiPath          = "/api"
	identityToken    = env("ACROCUIT_IDENTITY_TOKEN", "")
	spaceGroupId     = env("ACROCUIT_SPACE_GROUP_ID", "1")
	spaceId          = env("ACROCUIT_SPACE_ID", "1")
	switchPanelId    = env("ACROCUIT_SWITCH_PANEL_ID", "1")
	switchId         = env("ACROCUIT_SWITCH_ID", "1")
	upstreamSwitchId = env("ACROCUIT_UPSTREAM_SWITCH_ID", "2")
	deviceId         = env("ACROCUIT_DEVICE_ID", "1")
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func call(t *testing.T, method, path string, body any) {
	t.Helper()

	var reader io.Reader
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, host+apiPath+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if identityToken != "" {
		req.Header.Set("X-Forwarded-Identity", identityToken)
	}

	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Skipf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	t.Logf("%s %s %s\n-> %d %s", method, path, payload, res.StatusCode, raw)
}

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func TestGetMeta(t *testing.T) {
	call(t, http.MethodGet, "/meta", nil)
}

func TestGetSpaceGroups(t *testing.T) {
	call(t, http.MethodGet, "/space-groups", nil)
}

func TestAddSpaceGroup(t *testing.T) {
	call(t, http.MethodPost, "/space-groups", map[string]any{"name": "space group"})
}

func TestGetSpaceGroup(t *testing.T) {
	call(t, http.MethodGet, "/space-groups/"+spaceGroupId, nil)
}

func TestSetSpaceGroup(t *testing.T) {
	call(t, http.MethodPut, "/space-groups/"+spaceGroupId, map[string]any{"name": "space group"})
}

func TestDelSpaceGroup(t *testing.T) {
	call(t, http.MethodDelete, "/space-groups/"+spaceGroupId, nil)
}

func TestGetSpaces(t *testing.T) {
	call(t, http.MethodGet, "/space-groups/"+spaceGroupId+"/spaces", nil)
}

func TestAddSpace(t *testing.T) {
	call(t, http.MethodPost, "/space-groups/"+spaceGroupId+"/spaces", map[string]any{"name": "space"})
}

func TestGetSpace(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId, nil)
}

func TestSetSpace(t *testing.T) {
	call(t, http.MethodPut, "/spaces/"+spaceId, map[string]any{"name": "space"})
}

func TestDelSpace(t *testing.T) {
	call(t, http.MethodDelete, "/spaces/"+spaceId, nil)
}

func TestSetSpaceOrder(t *testing.T) {
	call(t, http.MethodPatch, "/spaces/"+spaceId+"/order", map[string]any{"display_order": 1})
}

func TestSetSpaceBackgroundImage(t *testing.T) {
	call(t, http.MethodPut, "/spaces/"+spaceId+"/background-image", nil)
}

func TestGetSpaceBackgroundImage(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId+"/background-image", nil)
}

func TestDelSpaceBackgroundImage(t *testing.T) {
	call(t, http.MethodDelete, "/spaces/"+spaceId+"/background-image", nil)
}

func TestGetSwitchPanels(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId+"/switch-panels", nil)
}

func TestAddSwitchPanel(t *testing.T) {
	call(t, http.MethodPost, "/spaces/"+spaceId+"/switch-panels", map[string]any{"name": "switch panel", "position": position{}, "attributes": map[string]any{}})
}

func TestGetDevices(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId+"/devices", nil)
}

func TestAddDevice(t *testing.T) {
	call(t, http.MethodPost, "/spaces/"+spaceId+"/devices", map[string]any{"name": "device", "position": position{}, "attributes": map[string]any{}})
}

func TestGetSwitchPanel(t *testing.T) {
	call(t, http.MethodGet, "/switch-panels/"+switchPanelId, nil)
}

func TestSetSwitchPanel(t *testing.T) {
	call(t, http.MethodPut, "/switch-panels/"+switchPanelId, map[string]any{"name": "switch panel", "attributes": map[string]any{}})
}

func TestDelSwitchPanel(t *testing.T) {
	call(t, http.MethodDelete, "/switch-panels/"+switchPanelId, nil)
}

func TestSetSwitchPanelPosition(t *testing.T) {
	call(t, http.MethodPut, "/switch-panels/"+switchPanelId+"/position", map[string]any{"position": position{}})
}

func TestGetSwitches(t *testing.T) {
	call(t, http.MethodGet, "/switch-panels/"+switchPanelId+"/switches", nil)
}

func TestAddSwitch(t *testing.T) {
	call(t, http.MethodPost, "/switch-panels/"+switchPanelId+"/switches", map[string]any{"name": "switch", "attributes": map[string]any{}})
}

func TestGetSwitch(t *testing.T) {
	call(t, http.MethodGet, "/switches/"+switchId, nil)
}

func TestGetSwitchDownstream(t *testing.T) {
	call(t, http.MethodGet, "/switches/"+switchId+"/downstream", nil)
}

func TestSetSwitch(t *testing.T) {
	call(t, http.MethodPut, "/switches/"+switchId, map[string]any{"name": "switch", "attributes": map[string]any{}})
}

func TestDelSwitch(t *testing.T) {
	call(t, http.MethodDelete, "/switches/"+switchId, nil)
}

func TestSetSwitchOrder(t *testing.T) {
	call(t, http.MethodPatch, "/switches/"+switchId+"/order", map[string]any{"display_order": 1})
}

func TestGetSwitchUpstreams(t *testing.T) {
	call(t, http.MethodGet, "/switches/"+switchId+"/upstreams", nil)
}

func TestAddSwitchUpstream(t *testing.T) {
	call(t, http.MethodPost, "/switches/"+switchId+"/upstreams", map[string]any{"upstream_switch_id": json.RawMessage(upstreamSwitchId), "attributes": map[string]any{}})
}

func TestSetSwitchUpstream(t *testing.T) {
	call(t, http.MethodPut, "/switches/"+switchId+"/upstreams/"+upstreamSwitchId, map[string]any{"attributes": map[string]any{}})
}

func TestDelSwitchUpstream(t *testing.T) {
	call(t, http.MethodDelete, "/switches/"+switchId+"/upstreams/"+upstreamSwitchId, nil)
}

func TestGetDevice(t *testing.T) {
	call(t, http.MethodGet, "/devices/"+deviceId, nil)
}

func TestSetDevice(t *testing.T) {
	call(t, http.MethodPut, "/devices/"+deviceId, map[string]any{"name": "device", "position": position{}})
}

func TestDelDevice(t *testing.T) {
	call(t, http.MethodDelete, "/devices/"+deviceId, nil)
}

func TestSetDevicePosition(t *testing.T) {
	call(t, http.MethodPut, "/devices/"+deviceId+"/position", map[string]any{"position": position{}})
}

func TestGetDeviceUpstream(t *testing.T) {
	call(t, http.MethodGet, "/devices/"+deviceId+"/upstream", nil)
}

func TestGetDeviceSwitches(t *testing.T) {
	call(t, http.MethodGet, "/devices/"+deviceId+"/switches", nil)
}

func TestAddDeviceSwitch(t *testing.T) {
	call(t, http.MethodPost, "/devices/"+deviceId+"/switches", map[string]any{"switch_id": json.RawMessage(switchId), "attributes": map[string]any{}})
}

func TestSetDeviceSwitch(t *testing.T) {
	call(t, http.MethodPut, "/devices/"+deviceId+"/switches/"+switchId, map[string]any{"attributes": map[string]any{}})
}

func TestDelDeviceSwitch(t *testing.T) {
	call(t, http.MethodDelete, "/devices/"+deviceId+"/switches/"+switchId, nil)
}
