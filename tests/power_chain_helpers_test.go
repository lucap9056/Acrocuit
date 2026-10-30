package test

import (
	"acrocuit/internal/models"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func powerChain(c *client, path string) models.PowerChainGraph {
	c.t.Helper()

	var chain models.PowerChainGraph
	req := newRequest(c.t, http.MethodGet, path, nil)
	status := c.Send(req, &chain)
	require.Equal(c.t, http.StatusOK, status, "GET %s", path)
	return chain
}

func downstream(c *client, switchId string) models.PowerChainGraph {
	c.t.Helper()
	return powerChain(c, fmt.Sprintf("/switches/%s/downstream", switchId))
}

func upstream(c *client, deviceId string) models.PowerChainGraph {
	c.t.Helper()
	return powerChain(c, fmt.Sprintf("/devices/%s/upstream", deviceId))
}

func idsOf(ids map[string]string, names ...string) []string {
	result := make([]string, len(names))
	for i, name := range names {
		result[i] = ids[name]
	}
	return result
}

func switchIds(switches []models.Switch) []string {
	result := make([]string, len(switches))
	for i, b := range switches {
		result[i] = string(b.Id)
	}
	return result
}

func chainSwitchIds(switches []models.ChainSwitch) []string {
	result := make([]string, len(switches))
	for i, b := range switches {
		result[i] = string(b.Id)
	}
	return result
}

func hasDeviceOnSwitch(chain models.PowerChainGraph, deviceId, switchId string) bool {
	return slices.ContainsFunc(chain.DeviceLinks, func(l models.DeviceLink) bool {
		return string(l.DeviceId) == deviceId && string(l.SwitchId) == switchId
	})
}

func findSwitchLink(chain models.PowerChainGraph, upstreamId, downstreamId string) (models.SwitchLink, bool) {
	index := slices.IndexFunc(chain.SwitchLinks, func(l models.SwitchLink) bool {
		return string(l.UpstreamSwitchId) == upstreamId && string(l.DownstreamSwitchId) == downstreamId
	})
	if index < 0 {
		return models.SwitchLink{}, false
	}
	return chain.SwitchLinks[index], true
}

func logTree(t *testing.T, rootId string, chain models.PowerChainGraph) {
	t.Helper()

	names := make(map[string]string, len(chain.Switches))
	for _, b := range chain.Switches {
		names[string(b.Id)] = b.Name
	}
	children := make(map[string][]string)
	for _, l := range chain.SwitchLinks {
		children[string(l.UpstreamSwitchId)] = append(children[string(l.UpstreamSwitchId)], string(l.DownstreamSwitchId))
	}
	devices := make(map[string]models.Device, len(chain.Devices))
	for _, d := range chain.Devices {
		devices[string(d.Id)] = d
	}
	attached := make(map[string][]models.Device)
	for _, l := range chain.DeviceLinks {
		attached[string(l.SwitchId)] = append(attached[string(l.SwitchId)], devices[string(l.DeviceId)])
	}

	var sb strings.Builder
	var walk func(id string, prefix string, last bool)
	walk = func(id string, prefix string, last bool) {
		branch, indent := "├─ ", "│  "
		if last {
			branch, indent = "└─ ", "   "
		}
		fmt.Fprintf(&sb, "%s%s[B] %s (#%s)\n", prefix, branch, names[id], id)

		kids, devs := children[id], attached[id]
		for i, d := range devs {
			leaf := "├─ "
			if i == len(devs)-1 && len(kids) == 0 {
				leaf = "└─ "
			}
			fmt.Fprintf(&sb, "%s%s[D] %s (#%s)\n", prefix+indent, leaf, d.Name, d.Id)
		}
		for i, k := range kids {
			walk(k, prefix+indent, i == len(kids)-1)
		}
	}
	walk(rootId, "", true)
	t.Logf("downstream of #%s:\n%s", rootId, sb.String())
}
