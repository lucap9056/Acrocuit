package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type spacesController struct {
	controller
}

const (
	paramSpaceID = "spaceId"
	routeSpaceID = ":" + paramSpaceID
)

func spacesHandlers(r *gin.RouterGroup, base controller) {
	ctr := &spacesController{base}

	r.GET(routeSpaceID, ctr.GetSpace)
	r.PUT(routeSpaceID, requireUser, ctr.SetSpace)
	r.DELETE(routeSpaceID, requireUser, ctr.DelSpace)
	r.PATCH(routeSpaceID+"/order", requireUser, ctr.SetSpaceOrder)

	r.GET(routeSpaceID+"/background-image", ctr.GetBackgroundImage)
	r.PUT(routeSpaceID+"/background-image", requireUser, ctr.SetBackgroundImage)
	r.DELETE(routeSpaceID+"/background-image", requireUser, ctr.DelBackgroundImage)

	r.GET(routeSpaceID+"/switch-panels", ctr.GetSwitchPanels)
	r.POST(routeSpaceID+"/switch-panels", requireUser, ctr.AddSwitchPanel)

	r.GET(routeSpaceID+"/devices", ctr.GetDevices)
	r.POST(routeSpaceID+"/devices", requireUser, ctr.AddDevice)
}

func (ctr *spacesController) GetSpace(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	space, err := ctr.Spaces.GetSpace(c.Request.Context(), userEmail, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

type setSpaceRequest struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=255"`
}

func (ctr *spacesController) SetSpace(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}

	var req setSpaceRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := ctr.Spaces.SetSpace(c.Request.Context(), userEmail, spaceId, req.Name)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}

func (ctr *spacesController) DelSpace(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}

	if err := ctr.Spaces.DelSpace(c.Request.Context(), userEmail, spaceId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

type setOrderRequest struct {
	DisplayOrder int `json:"display_order" binding:"required"`
}

func (ctr *spacesController) SetSpaceOrder(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}

	var req setOrderRequest
	if !bindJSON(c, &req) {
		return
	}

	spaces, err := ctr.Spaces.MoveSpace(c.Request.Context(), userEmail, spaceId, req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, spaces)
}

func (ctr *spacesController) GetSwitchPanels(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	switchPanels, err := ctr.SwitchPanels.GetSwitchPanels(c.Request.Context(), userEmail, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switchPanels)
}

type createSwitchPanelRequest struct {
	Name       string            `json:"name" binding:"required,min=1,max=255"`
	Position   *models.Position  `json:"position"`
	Attributes models.Attributes `json:"attributes"`
}

func (ctr *spacesController) AddSwitchPanel(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}

	var req createSwitchPanelRequest
	if !bindJSON(c, &req) {
		return
	}

	switchPanel, err := ctr.SwitchPanels.AddSwitchPanel(c.Request.Context(), userEmail, spaceId, req.Name, req.Position, req.Attributes.OrEmpty())
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switchPanel)
}

func (ctr *spacesController) GetDevices(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}
	include := c.Query("include")

	devices, err := ctr.Devices.GetDevices(c.Request.Context(), userEmail, spaceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, devices)
}

type createDeviceRequest struct {
	Name       string            `json:"name" binding:"required,min=1,max=255"`
	Position   *models.Position  `json:"position"`
	Attributes models.Attributes `json:"attributes"`
}

func (ctr *spacesController) AddDevice(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceId, ok := uuidParam[models.SpaceId](c, paramSpaceID)
	if !ok {
		return
	}

	var req createDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	device, err := ctr.Devices.AddDevice(c.Request.Context(), userEmail, spaceId, req.Name, req.Position, req.Attributes.OrEmpty())
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, device)
}
