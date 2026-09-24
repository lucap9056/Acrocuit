package handlers

import (
	"acrocuit/internal/response"
	"acrocuit/internal/storage"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func intParam(c *gin.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("invalid %s", name))
		return 0, false
	}
	return v, true
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func handleStorageErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found")
	case errors.Is(err, storage.ErrConflict):
		response.Error(c, http.StatusConflict, "conflict")
	case errors.Is(err, storage.ErrSelfReference), errors.Is(err, storage.ErrCycleDetected):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		internalError(c, err)
	}
}

func internalError(c *gin.Context, err error) {
	_ = c.Error(err)
	response.Error(c, http.StatusInternalServerError, "internal server error")
}
