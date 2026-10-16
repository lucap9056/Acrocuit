//go:build !noweb

package handlers

import (
	"acrocuit/internal/response"
	"acrocuit/web"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	webDistDir      = "dist"
	webIndexFile    = "index.html"
	webAssetsPrefix = "assets/"
	apiPathPrefix   = "/api/"
)

type webController struct{}

func webHandlers(g *gin.Engine) {
	ctr := &webController{}
	g.NoRoute(ctr.Serve)
}

func (ctr *webController) Serve(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, apiPathPrefix) {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}

	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		response.Error(c, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name, ok := ctr.resolve(c.Request.URL.Path)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	file, err := web.Dist.Open(path.Join(webDistDir, name))
	if err != nil {
		internalError(c, err)
		return
	}
	defer file.Close()

	content, ok := file.(io.ReadSeeker)
	if !ok {
		internalError(c, fs.ErrInvalid)
		return
	}

	if strings.HasPrefix(name, webAssetsPrefix) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}
	c.Header("X-Content-Type-Options", "nosniff")

	http.ServeContent(c.Writer, c.Request, name, time.Time{}, content)
}

func (ctr *webController) resolve(urlPath string) (string, bool) {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")

	if name != "" && ctr.isFile(name) {
		return name, true
	}

	if index := path.Join(name, webIndexFile); ctr.isFile(index) {
		return index, true
	}

	if path.Ext(name) != "" {
		return "", false
	}

	return webIndexFile, ctr.isFile(webIndexFile)
}

func (ctr *webController) isFile(name string) bool {
	info, err := fs.Stat(web.Dist, path.Join(webDistDir, name))
	return err == nil && !info.IsDir()
}
