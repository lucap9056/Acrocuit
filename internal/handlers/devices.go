package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type devicesController struct {
	controller
}

const (
	paramDeviceID = "deviceId"
	routeDeviceID = ":" + paramDeviceID
)

func devicesHandlers(r *gin.RouterGroup, base controller) {
	ctr := &devicesController{base}

	r.GET(routeDeviceID, ctr.GetDevice)
	r.PUT(routeDeviceID, requireUser, ctr.SetDevice)
	r.PUT(routeDeviceID+"/position", requireUser, ctr.SetDevicePosition)
	r.DELETE(routeDeviceID, requireUser, ctr.DelDevice)

	r.GET(routeDeviceID+"/switches", ctr.GetDeviceSwitches)
	r.POST(routeDeviceID+"/switches", requireUser, ctr.AddDeviceSwitch)
	r.PUT(routeDeviceID+"/switches/"+routeSwitchID, requireUser, ctr.SetDeviceSwitch)
	r.DELETE(routeDeviceID+"/switches/"+routeSwitchID, requireUser, ctr.DelDeviceSwitch)

	r.GET(routeDeviceID+"/upstream", ctr.GetDeviceUpstream)
}

func (ctr *devicesController) GetDevice(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	include := c.Query("include")

	device, err := ctr.Devices.GetDevice(c.Request.Context(), userEmail, deviceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, device)
}

type setDeviceRequest struct {
	Name       *string           `json:"name" binding:"omitempty,min=1,max=255"`
	Position   *models.Position  `json:"position"`
	Attributes models.Attributes `json:"attributes"`
}

func (ctr *devicesController) SetDevice(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	var req setDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	device, err := ctr.Devices.SetDevice(c.Request.Context(), userEmail, deviceId, req.Name, req.Position, req.Attributes)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, device)
}

func (ctr *devicesController) SetDevicePosition(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	var req setPositionRequest
	if !bindJSON(c, &req) {
		return
	}

	device, err := ctr.Devices.SetDevicePosition(c.Request.Context(), userEmail, deviceId, *req.Position)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, device)
}

func (ctr *devicesController) DelDevice(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	if err := ctr.Devices.DelDevice(c.Request.Context(), userEmail, deviceId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *devicesController) GetDeviceSwitches(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	include := c.Query("include")

	switches, err := ctr.Devices.GetDeviceSwitches(c.Request.Context(), userEmail, deviceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switches)
}

type addDeviceSwitchRequest struct {
	SwitchId   int               `json:"switch_id" binding:"required"`
	Attributes models.Attributes `json:"attributes"`
}

func (ctr *devicesController) AddDeviceSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	var req addDeviceSwitchRequest
	if !bindJSON(c, &req) {
		return
	}

	link, err := ctr.Devices.AddDeviceSwitch(c.Request.Context(), userEmail, deviceId, req.SwitchId, req.Attributes.OrEmpty())
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, link)
}

func (ctr *devicesController) SetDeviceSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	var req setAttributesRequest
	if !bindJSON(c, &req) {
		return
	}

	link, err := ctr.Devices.SetDeviceSwitch(c.Request.Context(), userEmail, deviceId, switchId, req.Attributes)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, link)
}

func (ctr *devicesController) DelDeviceSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	if err := ctr.Devices.DelDeviceSwitch(c.Request.Context(), userEmail, deviceId, switchId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *devicesController) GetDeviceUpstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	chain, err := ctr.Devices.GetDeviceUpstream(c.Request.Context(), userEmail, deviceId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, chain)
}
