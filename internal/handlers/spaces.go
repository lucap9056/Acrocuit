package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"acrocuit/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type spacesController struct {
	*storage.Storage
}

const (
	paramSpaceID = "spaceId"
	routeSpaceID = ":" + paramSpaceID
)

func spacesHandlers(r *gin.RouterGroup, s *storage.Storage) {
	ctr := &spacesController{s}

	r.GET(routeSpaceID, ctr.GetSpace)
	r.PUT(routeSpaceID, ctr.SetSpace)
	r.DELETE(routeSpaceID, ctr.DelSpace)
	r.PATCH(routeSpaceID+"/order", ctr.SetSpaceOrder)

	r.GET(routeSpaceID+"/breaker-groups", ctr.GetBreakerGroups)
	r.POST(routeSpaceID+"/breaker-groups", ctr.AddBreakerGroup)

	r.GET(routeSpaceID+"/devices", ctr.GetDevices)
	r.POST(routeSpaceID+"/devices", ctr.AddDevice)
}

func (ctr *spacesController) GetSpace(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	space, err := ctr.Spaces.GetSpace(c.Request.Context(), userId, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

type setSpaceRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=255"`
	DisplayOrder *int    `json:"display_order"`
}

func (ctr *spacesController) SetSpace(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	var req setSpaceRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := ctr.Spaces.SetSpace(c.Request.Context(), userId, spaceId, req.Name, req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

func (ctr *spacesController) DelSpace(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	if err := ctr.Spaces.DelSpace(c.Request.Context(), userId, spaceId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

type setOrderRequest struct {
	DisplayOrder int `json:"display_order" binding:"required"`
}

func (ctr *spacesController) SetSpaceOrder(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	var req setOrderRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := ctr.Spaces.SetSpace(c.Request.Context(), userId, spaceId, nil, &req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

func (ctr *spacesController) GetBreakerGroups(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	breakerGroups, err := ctr.BreakerGroups.GetBreakerGroups(c.Request.Context(), userId, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerGroups)
}

type createBreakerGroupRequest struct {
	Name     string           `json:"name" binding:"required,min=1,max=255"`
	Position *models.Position `json:"position"`
}

func (ctr *spacesController) AddBreakerGroup(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	var req createBreakerGroupRequest
	if !bindJSON(c, &req) {
		return
	}

	breakerGroup, err := ctr.BreakerGroups.AddBreakerGroup(c.Request.Context(), userId, spaceId, req.Name, req.Position)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerGroup)
}

func (ctr *spacesController) GetDevices(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	devices, err := ctr.Devices.GetDevices(c.Request.Context(), userId, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, devices)
}

type createDeviceRequest struct {
	Name     string           `json:"name" binding:"required,min=1,max=255"`
	Position *models.Position `json:"position"`
}

func (ctr *spacesController) AddDevice(c *gin.Context) {
	userId := auth.UserID(c)
	spaceId, ok := intParam(c, paramSpaceID)
	if !ok {
		return
	}

	var req createDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	device, err := ctr.Devices.AddDevice(c.Request.Context(), userId, spaceId, req.Name, req.Position)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, device)
}
