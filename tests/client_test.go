package test

import (
	"acrocuit/internal/response"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type client struct {
	t             *testing.T
	http          *http.Client
	base          *url.URL
	identityToken string
	accessToken   string
}

func newClient(t *testing.T) *client {
	t.Helper()
	base, err := url.Parse(host + apiPath)
	require.NoError(t, err, "parse base url")
	c := &client{
		t:             t,
		http:          &http.Client{Timeout: 10 * time.Second},
		base:          base,
		identityToken: identityToken,
		accessToken:   accessToken,
	}
	c.requireServer()
	return c
}

func (c *client) requireServer() {
	c.t.Helper()
	res, err := c.http.Get(c.base.String())
	if err != nil {
		c.t.Fatalf("server unreachable at %s: %v", c.base, err)
	}
	res.Body.Close()
}

func (c *client) with(t *testing.T) *client {
	clone := *c
	clone.t = t
	return &clone
}

func (c *client) do(req *http.Request) (int, http.Header, []byte) {
	c.t.Helper()

	target := c.base.JoinPath(req.URL.Path)
	target.RawQuery = req.URL.RawQuery
	req.URL = target
	if c.identityToken != "" {
		req.Header.Set("X-Forwarded-Identity", c.identityToken)
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	res, err := c.http.Do(req)
	require.NoError(c.t, err, "%s %s", req.Method, req.URL.Path)
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	require.NoError(c.t, err, "%s %s: read body", req.Method, req.URL.Path)
	if res.StatusCode >= http.StatusBadRequest {
		c.t.Logf("%s %s -> %d %s", req.Method, req.URL.Path, res.StatusCode, raw)
	}
	return res.StatusCode, res.Header, raw
}

func (c *client) Send(req *http.Request, data any) int {
	c.t.Helper()

	status, _, raw := c.do(req)
	if data != nil && len(raw) > 0 {
		err := json.Unmarshal(raw, &response.Response[any]{Data: data})
		require.NoError(c.t, err, "%s %s: decode response (%s)", req.Method, req.URL.Path, raw)
	}
	return status
}

func (c *client) BackgroundImage(spaceId string, etag string) (int, http.Header, []byte) {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/spaces/%s/background-image", spaceId), nil)
	require.NoError(c.t, err, "new request")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	return c.do(req)
}

func newRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err, "marshal body")
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, path, reader)
	require.NoError(t, err, "new request")
	req.Header.Set("Content-Type", "application/json")
	return req
}
