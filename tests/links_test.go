package test

import (
	"acrocuit/internal/models"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLinks(t *testing.T, c *client, b *building) {
	t.Run("SwitchUpstreamRejectsCycle", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct{ sw, upstream string }{
			{"switch1", "switch10"},
			{"switch3", "switch11"},
			{"switch10", "switch10"},
		} {
			path := fmt.Sprintf("/switches/%s/upstreams", b.switches[tc.sw])
			body := map[string]any{"upstream_switch_id": b.switches[tc.upstream]}

			req := newRequest(t, http.MethodPost, path, body)
			assert.Equal(t, http.StatusBadRequest, c.Send(req, nil), "add upstream %s to %s", tc.upstream, tc.sw)
		}
	})

	t.Run("SwitchAcceptsMultipleUpstreams", func(t *testing.T) {
		c := c.with(t)
		path := fmt.Sprintf("/switches/%s/upstreams", b.switches["switch9"])
		body := map[string]any{"upstream_switch_id": b.switches["switch6"]}

		req := newRequest(t, http.MethodPost, path, body)
		assert.Equal(t, http.StatusConflict, c.Send(req, nil), "link switch9 to switch6 again")

		var upstreams []models.Switch
		req = newRequest(t, http.MethodGet, path, nil)
		require.Equal(t, http.StatusOK, c.Send(req, &upstreams), "GET %s", path)
		assert.ElementsMatch(t, idsOf(b.switches, "switch5", "switch6"), switchIds(upstreams), "upstreams of switch9")
	})

	t.Run("AttachDeviceRejectsDuplicate", func(t *testing.T) {
		c := c.with(t)
		path := fmt.Sprintf("/devices/%s/switches", b.devices["device9"])
		body := map[string]any{"switch_id": b.switches["switch10"]}

		req := newRequest(t, http.MethodPost, path, body)
		assert.Equal(t, http.StatusConflict, c.Send(req, nil), "attach device9 to switch10 again")
	})
}
