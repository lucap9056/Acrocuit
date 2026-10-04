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
	r.PUT(routeDeviceID, ctr.SetDevice)
	r.PUT(routeDeviceID+"/position", ctr.SetDevicePosition)
	r.DELETE(routeDeviceID, ctr.DelDevice)

	r.GET(routeDeviceID+"/breakers", ctr.GetDeviceBreakers)
	r.POST(routeDeviceID+"/breakers", ctr.AddDeviceBreaker)
	r.DELETE(routeDeviceID+"/breakers/"+routeBreakerID, ctr.DelDeviceBreaker)

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
	Name     *string          `json:"name" binding:"omitempty,min=1,max=255"`
	Position *models.Position `json:"position"`
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

	device, err := ctr.Devices.SetDevice(c.Request.Context(), userEmail, deviceId, req.Name, req.Position)
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

func (ctr *devicesController) GetDeviceBreakers(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	include := c.Query("include")

	breakers, err := ctr.Devices.GetDeviceBreakers(c.Request.Context(), userEmail, deviceId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakers)
}

type addDeviceBreakerRequest struct {
	BreakerId int `json:"breaker_id" binding:"required"`
}

func (ctr *devicesController) AddDeviceBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}

	var req addDeviceBreakerRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.Devices.AddDeviceBreaker(c.Request.Context(), userEmail, deviceId, req.BreakerId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *devicesController) DelDeviceBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	deviceId, ok := intParam(c, paramDeviceID)
	if !ok {
		return
	}
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	if err := ctr.Devices.DelDeviceBreaker(c.Request.Context(), userEmail, deviceId, breakerId); err != nil {
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

	breakers, err := ctr.Devices.GetDeviceUpstream(c.Request.Context(), userEmail, deviceId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakers)
}
