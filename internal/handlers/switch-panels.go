package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type switchPanelsController struct {
	controller
}

const (
	paramSwitchPanelID = "switchPanelId"
	routeSwitchPanelID = ":" + paramSwitchPanelID
)

func switchPanelsHandlers(r *gin.RouterGroup, base controller) {
	ctr := &switchPanelsController{base}

	r.GET(routeSwitchPanelID, ctr.GetSwitchPanel)
	r.PUT(routeSwitchPanelID, ctr.SetSwitchPanel)
	r.PUT(routeSwitchPanelID+"/position", ctr.SetSwitchPanelPosition)
	r.DELETE(routeSwitchPanelID, ctr.DelSwitchPanel)

	r.GET(routeSwitchPanelID+"/switches", ctr.GetSwitches)
	r.POST(routeSwitchPanelID+"/switches", ctr.AddSwitch)
}

func (ctr *switchPanelsController) GetSwitchPanel(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}
	include := c.Query("include")

	switchPanel, err := ctr.SwitchPanels.GetSwitchPanel(c.Request.Context(), userEmail, switchPanelId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switchPanel)
}

type setSwitchPanelRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (ctr *switchPanelsController) SetSwitchPanel(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}

	var req setSwitchPanelRequest
	if !bindJSON(c, &req) {
		return
	}

	switchPanel, err := ctr.SwitchPanels.SetSwitchPanel(c.Request.Context(), userEmail, switchPanelId, req.Name)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switchPanel)
}

type setPositionRequest struct {
	Position *models.Position `json:"position" binding:"required"`
}

func (ctr *switchPanelsController) SetSwitchPanelPosition(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}

	var req setPositionRequest
	if !bindJSON(c, &req) {
		return
	}

	switchPanel, err := ctr.SwitchPanels.SetSwitchPanelPosition(c.Request.Context(), userEmail, switchPanelId, *req.Position)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switchPanel)
}

func (ctr *switchPanelsController) DelSwitchPanel(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}

	if err := ctr.SwitchPanels.DelSwitchPanel(c.Request.Context(), userEmail, switchPanelId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *switchPanelsController) GetSwitches(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}
	include := c.Query("include")

	switches, err := ctr.Switches.GetSwitches(c.Request.Context(), userEmail, switchPanelId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switches)
}

type createSwitchRequest struct {
	Name             string `json:"name" binding:"required,min=1,max=255"`
	UpstreamSwitchId int    `json:"upstream_switch_id"`
}

func (ctr *switchPanelsController) AddSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchPanelId, ok := intParam(c, paramSwitchPanelID)
	if !ok {
		return
	}

	var req createSwitchRequest
	if !bindJSON(c, &req) {
		return
	}

	var upstreamSwitchId *int
	if req.UpstreamSwitchId != 0 {
		upstreamSwitchId = &req.UpstreamSwitchId
	}

	ctx := c.Request.Context()
	sw, err := ctr.Switches.AddSwitch(ctx, userEmail, switchPanelId, req.Name, upstreamSwitchId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, sw)
}
