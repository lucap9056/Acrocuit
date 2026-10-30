package test

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeleteSpaceRemovesCrossSpaceReferences(t *testing.T, c *client, b *building) {
	root := b.switches["switch1"]
	shared := b.devices["device9"]

	removedPath := fmt.Sprintf("/spaces/%s", b.spaces["space5"])
	req := newRequest(t, http.MethodDelete, removedPath, nil)
	require.Equal(t, http.StatusOK, c.Send(req, nil), "delete space5")

	for _, path := range []string{
		removedPath,
		fmt.Sprintf("/switch-panels/%s", b.switchPanels["group5"]),
		fmt.Sprintf("/switches/%s", b.switches["switch8"]),
		fmt.Sprintf("/switches/%s", b.switches["switch11"]),
		fmt.Sprintf("/devices/%s", b.devices["device10"]),
		fmt.Sprintf("/devices/%s", b.devices["device11"]),
	} {
		req := newRequest(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusNotFound, c.Send(req, nil), "GET %s after deleting space5", path)
	}

	want := idsOf(b.switches, "switch10", "switch7", "switch3", "switch1")
	assert.ElementsMatch(t, want, chainSwitchIds(upstream(c, shared).Switches), "upstream of device9 after deleting space5")

	remaining := slices.DeleteFunc(b.switchNames(), func(name string) bool {
		return name == "switch8" || name == "switch11"
	})
	chain := downstream(c, root)
	logTree(t, root, chain)

	assert.ElementsMatch(t, idsOf(b.switches, remaining...), chainSwitchIds(chain.Switches), "downstream switches after deleting space5")
	assert.True(t, hasDeviceOnSwitch(chain, shared, b.switches["switch10"]), "device9 should remain under switch10")
	assert.True(t, hasDeviceOnSwitch(chain, b.devices["device3"], b.switches["switch5"]), "device3 should be unaffected")
}

func testDeleteSpaceGroupCascades(t *testing.T, c *client, b *building) {
	req := newRequest(t, http.MethodDelete, b.spaceGroupPath, nil)
	require.Equal(t, http.StatusOK, c.Send(req, nil), "delete space group")

	req = newRequest(t, http.MethodGet, b.spaceGroupPath, nil)
	assert.Equal(t, http.StatusNotFound, c.Send(req, nil), "GET %s after delete", b.spaceGroupPath)
	for _, tc := range []struct {
		prefix string
		ids    map[string]string
	}{
		{"/spaces", b.spaces},
		{"/switch-panels", b.switchPanels},
		{"/switches", b.switches},
		{"/devices", b.devices},
	} {
		for _, id := range tc.ids {
			path := fmt.Sprintf("%s/%s", tc.prefix, id)
			req := newRequest(t, http.MethodGet, path, nil)
			assert.Equal(t, http.StatusNotFound, c.Send(req, nil), "GET %s after deleting space group", path)
		}
	}
}
