package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testIncludeRejectsUnknownColumn(t *testing.T, c *client, b *building) {
	for _, path := range []string{
		b.spaceGroupPath,
		fmt.Sprintf("/spaces/%s", b.spaces["space1"]),
		fmt.Sprintf("/spaces/%s/switch-panels", b.spaces["space1"]),
		fmt.Sprintf("/spaces/%s/devices", b.spaces["space1"]),
		fmt.Sprintf("/switches/%s", b.switches["switch1"]),
	} {
		req := newRequest(t, http.MethodGet, path+"?include=id,unknown", nil)
		assert.Equal(t, http.StatusBadRequest, c.Send(req, nil), "GET %s", path)
	}
}
