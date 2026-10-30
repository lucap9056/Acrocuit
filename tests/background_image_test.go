package test

import (
	"acrocuit/internal/models"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pngImage(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size)))
	require.NoError(t, err, "encode png")
	return buf.Bytes()
}

func backgroundImageRequest(t *testing.T, spaceId string, field string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, "background")
	require.NoError(t, err, "create form file")
	part.Write(data)
	writer.Close()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/spaces/%s/background-image", spaceId), &body)
	require.NoError(t, err, "new request")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func testSpaceBackgroundImage(t *testing.T, c *client, b *building) {
	spaceId := b.spaces["space2"]
	spacePath := fmt.Sprintf("/spaces/%s", spaceId)
	backgroundImagePath := spacePath + "/background-image"

	t.Run("RoundTrip", func(t *testing.T) {
		c := c.with(t)
		status, _, _ := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusNotFound, status, "GET background image before upload")

		var space models.Space
		req := newRequest(t, http.MethodGet, spacePath, nil)
		status = c.Send(req, &space)
		require.Equal(t, http.StatusOK, status, "GET %s", spacePath)
		createdAt := space.BackgroundImageUpdatedAt

		time.Sleep(5 * time.Millisecond)
		first := pngImage(t, 2)
		var uploaded models.Space
		req = backgroundImageRequest(t, spaceId, "image", first)
		status = c.Send(req, &uploaded)
		require.Equal(t, http.StatusOK, status, "upload background image")
		uploadedAt := uploaded.BackgroundImageUpdatedAt
		assert.Greater(t, uploadedAt, createdAt, "background_image_updated_at after upload")

		status, header, got := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusOK, status, "GET background image")
		assert.Equal(t, "image/png", header.Get("Content-Type"), "GET background image Content-Type")
		assert.Equal(t, first, got, "GET background image body")

		etag := header.Get("ETag")
		assert.Equal(t, fmt.Sprintf(`"%d"`, uploadedAt), etag, "ETag")
		status, _, _ = c.BackgroundImage(spaceId, etag)
		assert.Equal(t, http.StatusNotModified, status, "GET with matching If-None-Match")

		second := pngImage(t, 4)
		req = backgroundImageRequest(t, spaceId, "image", second)
		status = c.Send(req, nil)
		require.Equal(t, http.StatusOK, status, "replace background image")
		_, _, got = c.BackgroundImage(spaceId, "")
		assert.Equal(t, second, got, "GET after replace")

		for _, tc := range []struct {
			method string
			want   int
		}{
			{http.MethodDelete, http.StatusOK},
			{http.MethodGet, http.StatusNotFound},
			{http.MethodDelete, http.StatusNotFound},
		} {
			req := newRequest(t, tc.method, backgroundImagePath, nil)
			status := c.Send(req, nil)
			assert.Equal(t, tc.want, status, "%s background image", tc.method)
		}
	})

	t.Run("RejectsInvalidUpload", func(t *testing.T) {
		c := c.with(t)
		for _, tc := range []struct {
			space string
			field string
			data  []byte
			want  int
		}{
			{spaceId, "image", []byte("plain text, not an image"), http.StatusUnsupportedMediaType},
			{spaceId, "file", pngImage(t, 2), http.StatusBadRequest},
			{spaceId, "image", make([]byte, 11<<20), http.StatusRequestEntityTooLarge},
			{uuid.Nil.String(), "image", pngImage(t, 2), http.StatusNotFound},
		} {
			req := backgroundImageRequest(t, tc.space, tc.field, tc.data)
			status := c.Send(req, nil)
			assert.Equal(t, tc.want, status, "upload %q (%d bytes) to space %s", tc.field, len(tc.data), tc.space)
		}
		status, _, _ := c.BackgroundImage(spaceId, "")
		assert.Equal(t, http.StatusNotFound, status, "GET background image after rejected uploads")
	})
}
