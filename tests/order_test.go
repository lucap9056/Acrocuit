package test

import (
	"acrocuit/internal/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeleteCompactsDisplayOrder(t *testing.T, c *client, _ *building) {
	spaceGroupPath := newSpaceGroup(t, c)
	first := newSpace(t, c, spaceGroupPath, "first")
	middle := newSpace(t, c, spaceGroupPath, "middle")
	last := newSpace(t, c, spaceGroupPath, "last")

	var spaces []models.Space
	req := newRequest(t, http.MethodDelete, "/spaces/"+middle, nil)
	require.Equal(t, http.StatusOK, c.Send(req, &spaces), "delete middle space")
	require.Equal(t, []string{first, last}, spaceIds(spaces), "spaces after deleting the middle one")
	assert.Equal(t, 2, spaces[1].DisplayOrder, "last space moves up")

	switchPanelId := newSwitchPanel(t, c, first, "panel")
	firstSwitch := newSwitch(t, c, switchPanelId, "first")
	secondSwitch := newSwitch(t, c, switchPanelId, "second")

	var switches []models.Switch
	req = newRequest(t, http.MethodDelete, "/switches/"+firstSwitch, nil)
	require.Equal(t, http.StatusOK, c.Send(req, &switches), "delete first switch")
	require.Equal(t, []string{secondSwitch}, switchIds(switches), "switches after deleting the first one")
	assert.Equal(t, 1, switches[0].DisplayOrder, "second switch moves up")

	req = newRequest(t, http.MethodDelete, "/switches/"+secondSwitch, nil)
	require.Equal(t, http.StatusOK, c.Send(req, &switches), "delete the last switch")
	assert.Empty(t, switches, "switches after deleting every switch")

	req = newRequest(t, http.MethodDelete, "/switches/"+secondSwitch, nil)
	assert.Equal(t, http.StatusNotFound, c.Send(req, nil), "delete an already deleted switch")
}
