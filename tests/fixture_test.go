package test

import (
	"acrocuit/internal/models"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type building struct {
	spaceGroupPath string
	spaces         map[string]string
	switchPanels   map[string]string
	switches       map[string]string
	devices        map[string]string
}

func (b *building) switchNames() []string {
	names := make([]string, 0, len(b.switches))
	for name := range b.switches {
		names = append(names, name)
	}
	return names
}

func newSpaceGroup(t *testing.T, c *client) string {
	t.Helper()

	var spaceGroup models.SpaceGroup
	req := newRequest(t, http.MethodPost, "/space-groups", map[string]any{"name": t.Name()})
	require.Equal(t, http.StatusOK, c.Send(req, &spaceGroup), "create space group")
	path := fmt.Sprintf("/space-groups/%s", spaceGroup.Id)
	t.Cleanup(func() {
		req := newRequest(t, http.MethodDelete, path, nil)
		c.Send(req, nil)
	})
	return path
}

func newSpace(t *testing.T, c *client, spaceGroupPath, name string) string {
	t.Helper()

	var space models.Space
	req := newRequest(t, http.MethodPost, spaceGroupPath+"/spaces", map[string]any{"name": name})
	require.Equal(t, http.StatusOK, c.Send(req, &space), "create space %s", name)
	return string(space.Id)
}

func newBuilding(t *testing.T, c *client) *building {
	t.Helper()

	b := &building{
		spaceGroupPath: newSpaceGroup(t, c),
		spaces:         make(map[string]string),
		switchPanels:   make(map[string]string),
		switches:       make(map[string]string),
		devices:        make(map[string]string),
	}

	for _, name := range []string{"space1", "space2", "space3", "space4", "space5"} {
		b.spaces[name] = newSpace(t, c, b.spaceGroupPath, name)
	}

	for _, tc := range []struct{ space, name string }{
		{"space1", "group1"},
		{"space2", "group2"},
		{"space3", "group3"},
		{"space4", "group4"},
		{"space5", "group5"},
	} {
		path := fmt.Sprintf("/spaces/%s/switch-panels", b.spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var switchPanel models.SwitchPanel
		req := newRequest(t, http.MethodPost, path, body)
		require.Equal(t, http.StatusOK, c.Send(req, &switchPanel), "create switch panel %s", tc.name)
		b.switchPanels[tc.name] = string(switchPanel.Id)
	}

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
		path := fmt.Sprintf("/switch-panels/%s/switches", b.switchPanels[tc.group])
		body := map[string]any{"name": tc.name}

		var sw models.Switch
		req := newRequest(t, http.MethodPost, path, body)
		require.Equal(t, http.StatusOK, c.Send(req, &sw), "create switch %s", tc.name)
		b.switches[tc.name] = string(sw.Id)

		if tc.upstream != "" {
			b.linkSwitch(t, c, tc.name, tc.upstream, map[string]any{"from": tc.upstream})
		}
	}
	b.linkSwitch(t, c, "switch9", "switch6", nil)

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
		path := fmt.Sprintf("/spaces/%s/devices", b.spaces[tc.space])
		body := map[string]any{"name": tc.name}

		var device models.Device
		req := newRequest(t, http.MethodPost, path, body)
		require.Equal(t, http.StatusOK, c.Send(req, &device), "create device %s", tc.name)
		b.devices[tc.name] = string(device.Id)
	}

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
		path := fmt.Sprintf("/devices/%s/switches", b.devices[tc.device])
		body := map[string]any{"switch_id": b.switches[tc.sw]}

		req := newRequest(t, http.MethodPost, path, body)
		require.Equal(t, http.StatusOK, c.Send(req, nil), "attach %s to %s", tc.device, tc.sw)
	}

	return b
}

func (b *building) linkSwitch(t *testing.T, c *client, name, upstream string, attributes map[string]any) {
	t.Helper()

	path := fmt.Sprintf("/switches/%s/upstreams", b.switches[name])
	body := map[string]any{"upstream_switch_id": b.switches[upstream]}
	if attributes != nil {
		body["attributes"] = attributes
	}
	req := newRequest(t, http.MethodPost, path, body)
	require.Equal(t, http.StatusOK, c.Send(req, nil), "link %s to %s", name, upstream)
}
