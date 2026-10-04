package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/response"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	backgroundImageField     = "image"
	maxBackgroundImageSize   = 10 << 20
	maxBackgroundRequestSize = maxBackgroundImageSize + 1<<20
)

var backgroundImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/gif":  true,
}

func (ctr *spacesController) GetBackgroundImage(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	image, err := ctr.Spaces.GetSpaceBackgroundImage(c.Request.Context(), userEmail, spaceId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	etag := fmt.Sprintf(`"%d"`, image.UpdatedAt)
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(http.StatusOK, image.ContentType, image.Data)
}

func (ctr *spacesController) SetBackgroundImage(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	data, ok := readBackgroundImage(c)
	if !ok {
		return
	}

	contentType := http.DetectContentType(data)
	if !backgroundImageTypes[contentType] {
		response.Error(c, http.StatusUnsupportedMediaType, "background image must be PNG, JPEG, WebP or GIF")
		return
	}

	space, err := ctr.Spaces.SetSpaceBackgroundImage(c.Request.Context(), userEmail, spaceId, contentType, data)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

func (ctr *spacesController) DelBackgroundImage(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	space, err := ctr.Spaces.DelSpaceBackgroundImage(c.Request.Context(), userEmail, spaceId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

func readBackgroundImage(c *gin.Context) ([]byte, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBackgroundRequestSize)

	header, err := c.FormFile(backgroundImageField)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.Error(c, http.StatusRequestEntityTooLarge, "background image must not exceed 10 MB")
			return nil, false
		}
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("multipart field %q is required", backgroundImageField))
		return nil, false
	}

	file, err := header.Open()
	if err != nil {
		internalError(c, err)
		return nil, false
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxBackgroundImageSize+1))
	if err != nil {
		internalError(c, err)
		return nil, false
	}
	if len(data) > maxBackgroundImageSize {
		response.Error(c, http.StatusRequestEntityTooLarge, "background image must not exceed 10 MB")
		return nil, false
	}

	return data, true
}
