package test

import (
	"acrocuit/internal/models"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPowerChain(t *testing.T, c *client, b *building) {
	root := b.switches["switch1"]
	left := b.switches["switch10"]
	right := b.switches["switch11"]
	shared := b.devices["device9"]

	t.Run("SwitchDownstreamIncludesDescendants", func(t *testing.T) {
		c := c.with(t)
		chain := downstream(c, root)
		logTree(t, root, chain)

		assert.ElementsMatch(t, idsOf(b.switches, b.switchNames()...), chainSwitchIds(chain.Switches), "downstream switches")
		for _, l := range chain.SwitchLinks {
			assert.Contains(t, chainSwitchIds(chain.Switches), string(l.UpstreamSwitchId), "downstream link upstream end")
			assert.Contains(t, chainSwitchIds(chain.Switches), string(l.DownstreamSwitchId), "downstream link downstream end")
		}
		assert.Len(t, chain.Devices, len(b.devices), "downstream devices")
		assert.True(t, hasDeviceOnSwitch(chain, shared, left), "device9 should appear under switch10")
		assert.True(t, hasDeviceOnSwitch(chain, shared, right), "device9 should appear under switch11")

		link, ok := findSwitchLink(chain, b.switches["switch2"], b.switches["switch4"])
		if assert.True(t, ok, "link switch2 -> switch4") {
			assert.JSONEq(t, `{"from":"switch2"}`, string(link.Attributes), "link switch2 -> switch4 attributes")
		}
	})

	t.Run("DeviceUpstreamFollowsEveryUpstream", func(t *testing.T) {
		c := c.with(t)
		chain := upstream(c, b.devices["device4"])
		want := idsOf(b.switches, "switch9", "switch5", "switch6", "switch2", "switch1")
		assert.ElementsMatch(t, want, chainSwitchIds(chain.Switches), "upstream of device4")
		_, ok := findSwitchLink(chain, b.switches["switch6"], b.switches["switch9"])
		assert.True(t, ok, "link switch6 -> switch9")
	})

	t.Run("DeviceUpstreamDeduplicatesSharedPath", func(t *testing.T) {
		c := c.with(t)
		want := idsOf(b.switches, "switch10", "switch7", "switch11", "switch8", "switch3", "switch1")
		chain := upstream(c, shared)
		assert.ElementsMatch(t, want, chainSwitchIds(chain.Switches), "upstream of device9")
		assert.Len(t, chain.SwitchLinks, len(want)-1, "upstream switch links of device9")
		linked := make([]string, len(chain.DeviceLinks))
		for i, l := range chain.DeviceLinks {
			linked[i] = string(l.SwitchId)
		}
		assert.ElementsMatch(t, []string{left, right}, linked, "device links of device9")
	})

	t.Run("IncludesPanelsAndPositions", func(t *testing.T) {
		c := c.with(t)
		panelPosition := models.Position{X: 1, Y: 2, Z: 3}
		devicePosition := models.Position{X: 4, Y: 5, Z: 6}

		path := fmt.Sprintf("/switch-panels/%s/position", b.switchPanels["group4"])
		req := newRequest(t, http.MethodPut, path, map[string]any{"position": panelPosition})
		require.Equal(t, http.StatusOK, c.Send(req, nil), "PUT %s", path)

		path = fmt.Sprintf("/devices/%s/position", b.devices["device8"])
		req = newRequest(t, http.MethodPut, path, map[string]any{"position": devicePosition})
		require.Equal(t, http.StatusOK, c.Send(req, nil), "PUT %s", path)

		assertPanel := func(sw models.ChainSwitch, source string) {
			t.Helper()
			panel := sw.SwitchPanel
			require.NotNil(t, panel, "%s: switch #%s panel", source, sw.Id)
			require.NotNil(t, panel.Space, "%s: switch #%s panel space", source, sw.Id)
			if string(panel.Id) == b.switchPanels["group4"] {
				assert.Equal(t, "group4", panel.Name, "%s: panel name", source)
				assert.Equal(t, b.spaces["space4"], string(panel.Space.Id), "%s: panel space", source)
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
			if string(d.Id) == b.devices["device8"] {
				assert.Equal(t, &devicePosition, d.Position, "downstream device8 position")
			} else {
				assert.Nil(t, d.Position, "downstream device #%s position", d.Id)
			}
		}

		for _, sw := range upstream(c, shared).Switches {
			assertPanel(sw, "upstream")
		}
	})
}
