package test

import (
	"acrocuit/internal/models"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeleteCompactsDisplayOrder(t *testing.T, c *client, _ *building) {
	spaceGroupPath := newSpaceGroup(t, c)
	spaceIds := []string{
		newSpace(t, c, spaceGroupPath, "first"),
		newSpace(t, c, spaceGroupPath, "middle"),
		newSpace(t, c, spaceGroupPath, "last"),
	}

	var spaces []models.Space
	req := newRequest(t, http.MethodDelete, "/spaces/"+spaceIds[1], nil)
	require.Equal(t, http.StatusOK, c.Send(req, &spaces), "delete middle space")
	require.Len(t, spaces, 2)
	for i, wantId := range []string{spaceIds[0], spaceIds[2]} {
		assert.Equal(t, wantId, string(spaces[i].Id), "space at index %d", i)
		assert.Equal(t, i+1, spaces[i].DisplayOrder, "display_order of space at index %d", i)
	}

	var switchPanel models.SwitchPanel
	req = newRequest(t, http.MethodPost, fmt.Sprintf("/spaces/%s/switch-panels", spaceIds[0]), map[string]any{"name": "panel"})
	require.Equal(t, http.StatusOK, c.Send(req, &switchPanel), "create switch panel")

	switchIds := make([]string, 0, 3)
	for _, name := range []string{"first", "second", "third"} {
		var sw models.Switch
		req := newRequest(t, http.MethodPost, fmt.Sprintf("/switch-panels/%s/switches", switchPanel.Id), map[string]any{"name": name})
		require.Equal(t, http.StatusOK, c.Send(req, &sw), "create switch %s", name)
		switchIds = append(switchIds, string(sw.Id))
	}

	var switches []models.Switch
	req = newRequest(t, http.MethodDelete, "/switches/"+switchIds[0], nil)
	require.Equal(t, http.StatusOK, c.Send(req, &switches), "delete first switch")
	require.Len(t, switches, 2)
	for i, wantId := range []string{switchIds[1], switchIds[2]} {
		assert.Equal(t, wantId, string(switches[i].Id), "switch at index %d", i)
		assert.Equal(t, i+1, switches[i].DisplayOrder, "display_order of switch at index %d", i)
	}

	for _, id := range switchIds[1:] {
		req := newRequest(t, http.MethodDelete, "/switches/"+id, nil)
		require.Equal(t, http.StatusOK, c.Send(req, &switches), "delete switch %s", id)
	}
	assert.Empty(t, switches, "switches after deleting every switch")

	req = newRequest(t, http.MethodDelete, "/switches/"+switchIds[0], nil)
	assert.Equal(t, http.StatusNotFound, c.Send(req, nil), "delete an already deleted switch")
}
