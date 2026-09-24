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
	host              = env("ACROCUIT_HOST", "http://localhost:8080")
	accessToken       = env("ACROCUIT_ACCESS_TOKEN", "")
	refreshToken      = env("ACROCUIT_REFRESH_TOKEN", "")
	username          = env("ACROCUIT_USERNAME", "user")
	password          = env("ACROCUIT_PASSWORD", "Sm@rtHome123")
	spaceGroupId      = env("ACROCUIT_SPACE_GROUP_ID", "1")
	spaceId           = env("ACROCUIT_SPACE_ID", "1")
	breakerGroupId    = env("ACROCUIT_BREAKER_GROUP_ID", "1")
	breakerId         = env("ACROCUIT_BREAKER_ID", "1")
	upstreamBreakerId = env("ACROCUIT_UPSTREAM_BREAKER_ID", "2")
	deviceId          = env("ACROCUIT_DEVICE_ID", "1")
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

	req, err := http.NewRequest(method, host+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
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

func TestRegister(t *testing.T) {
	call(t, http.MethodPost, "/auth/register", map[string]any{"username": username, "password": password})
}

func TestLogin(t *testing.T) {
	call(t, http.MethodPost, "/auth/login", map[string]any{"username": username, "password": password})
}

func TestRefresh(t *testing.T) {
	call(t, http.MethodPost, "/auth/refresh", map[string]any{"refresh_token": refreshToken})
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
	call(t, http.MethodPut, "/spaces/"+spaceId, map[string]any{"name": "space", "display_order": 1})
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

func TestGetBreakerGroups(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId+"/breaker-groups", nil)
}

func TestAddBreakerGroup(t *testing.T) {
	call(t, http.MethodPost, "/spaces/"+spaceId+"/breaker-groups", map[string]any{"name": "breaker group", "position": position{}})
}

func TestGetDevices(t *testing.T) {
	call(t, http.MethodGet, "/spaces/"+spaceId+"/devices", nil)
}

func TestAddDevice(t *testing.T) {
	call(t, http.MethodPost, "/spaces/"+spaceId+"/devices", map[string]any{"name": "device", "position": position{}})
}

func TestGetBreakerGroup(t *testing.T) {
	call(t, http.MethodGet, "/breaker-groups/"+breakerGroupId, nil)
}

func TestSetBreakerGroup(t *testing.T) {
	call(t, http.MethodPut, "/breaker-groups/"+breakerGroupId, map[string]any{"name": "breaker group"})
}

func TestDelBreakerGroup(t *testing.T) {
	call(t, http.MethodDelete, "/breaker-groups/"+breakerGroupId, nil)
}

func TestSetBreakerGroupPosition(t *testing.T) {
	call(t, http.MethodPut, "/breaker-groups/"+breakerGroupId+"/position", map[string]any{"position": position{}})
}

func TestGetBreakers(t *testing.T) {
	call(t, http.MethodGet, "/breaker-groups/"+breakerGroupId+"/breakers", nil)
}

func TestAddBreaker(t *testing.T) {
	call(t, http.MethodPost, "/breaker-groups/"+breakerGroupId+"/breakers", map[string]any{"name": "breaker", "upstream_breaker_id": json.RawMessage(upstreamBreakerId)})
}

func TestGetBreaker(t *testing.T) {
	call(t, http.MethodGet, "/breakers/"+breakerId, nil)
}

func TestGetBreakerDownstream(t *testing.T) {
	call(t, http.MethodGet, "/breakers/"+breakerId+"/downstream", nil)
}

func TestSetBreaker(t *testing.T) {
	call(t, http.MethodPut, "/breakers/"+breakerId, map[string]any{"name": "breaker", "display_order": 1})
}

func TestDelBreaker(t *testing.T) {
	call(t, http.MethodDelete, "/breakers/"+breakerId, nil)
}

func TestSetBreakerOrder(t *testing.T) {
	call(t, http.MethodPatch, "/breakers/"+breakerId+"/order", map[string]any{"display_order": 1})
}

func TestSetBreakerUpstream(t *testing.T) {
	call(t, http.MethodPatch, "/breakers/"+breakerId+"/upstream", map[string]any{"upstream_breaker_id": json.RawMessage(upstreamBreakerId)})
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

func TestGetDeviceBreakers(t *testing.T) {
	call(t, http.MethodGet, "/devices/"+deviceId+"/breakers", nil)
}

func TestAddDeviceBreaker(t *testing.T) {
	call(t, http.MethodPost, "/devices/"+deviceId+"/breakers", map[string]any{"breaker_id": json.RawMessage(breakerId)})
}

func TestDelDeviceBreaker(t *testing.T) {
	call(t, http.MethodDelete, "/devices/"+deviceId+"/breakers/"+breakerId, nil)
}
