package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type switchesController struct {
	controller
}

const (
	paramSwitchID         = "switchId"
	routeSwitchID         = ":" + paramSwitchID
	paramUpstreamSwitchID = "upstreamSwitchId"
	routeUpstreamSwitchID = ":" + paramUpstreamSwitchID
)

func switchesHandlers(r *gin.RouterGroup, base controller) {
	ctr := &switchesController{base}

	r.GET(routeSwitchID+"/downstream", ctr.GetSwitchDownstream)
	r.GET(routeSwitchID, ctr.GetSwitch)
	r.PUT(routeSwitchID, requireUser, ctr.SetSwitch)
	r.DELETE(routeSwitchID, requireUser, ctr.DelSwitch)
	r.PATCH(routeSwitchID+"/order", requireUser, ctr.SetSwitchOrder)

	r.GET(routeSwitchID+"/upstreams", ctr.GetSwitchUpstreams)
	r.POST(routeSwitchID+"/upstreams", requireUser, ctr.AddSwitchUpstream)
	r.PUT(routeSwitchID+"/upstreams/"+routeUpstreamSwitchID, requireUser, ctr.SetSwitchUpstream)
	r.DELETE(routeSwitchID+"/upstreams/"+routeUpstreamSwitchID, requireUser, ctr.DelSwitchUpstream)
}

type setAttributesRequest struct {
	Attributes models.Attributes `json:"attributes" binding:"required"`
}

func (ctr *switchesController) GetSwitchDownstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	chain, err := ctr.Switches.GetSwitchDownstream(c.Request.Context(), userEmail, switchId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, chain)
}

func (ctr *switchesController) GetSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}
	include := c.Query("include")

	sw, err := ctr.Switches.GetSwitch(c.Request.Context(), userEmail, switchId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, sw)
}

type setSwitchRequest struct {
	Name       *string           `json:"name" binding:"omitempty,min=1,max=255"`
	Attributes models.Attributes `json:"attributes"`
}

func (ctr *switchesController) SetSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	var req setSwitchRequest
	if !bindJSON(c, &req) {
		return
	}

	sw, err := ctr.Switches.SetSwitch(c.Request.Context(), userEmail, switchId, req.Name, req.Attributes)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, sw)
}

func (ctr *switchesController) DelSwitch(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	if err := ctr.Switches.DelSwitch(c.Request.Context(), userEmail, switchId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *switchesController) SetSwitchOrder(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	var req setOrderRequest
	if !bindJSON(c, &req) {
		return
	}

	switches, err := ctr.Switches.MoveSwitch(c.Request.Context(), userEmail, switchId, req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switches)
}

func (ctr *switchesController) GetSwitchUpstreams(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}
	include := c.Query("include")

	switches, err := ctr.Switches.GetSwitchUpstreams(c.Request.Context(), userEmail, switchId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, switches)
}

type addSwitchUpstreamRequest struct {
	UpstreamSwitchId int               `json:"upstream_switch_id" binding:"required"`
	Attributes       models.Attributes `json:"attributes"`
}

func (ctr *switchesController) AddSwitchUpstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}

	var req addSwitchUpstreamRequest
	if !bindJSON(c, &req) {
		return
	}

	link, err := ctr.Switches.AddSwitchUpstream(c.Request.Context(), userEmail, switchId, req.UpstreamSwitchId, req.Attributes.OrEmpty())
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, link)
}

func (ctr *switchesController) SetSwitchUpstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}
	upstreamSwitchId, ok := intParam(c, paramUpstreamSwitchID)
	if !ok {
		return
	}

	var req setAttributesRequest
	if !bindJSON(c, &req) {
		return
	}

	link, err := ctr.Switches.SetSwitchUpstream(c.Request.Context(), userEmail, switchId, upstreamSwitchId, req.Attributes)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, link)
}

func (ctr *switchesController) DelSwitchUpstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	switchId, ok := intParam(c, paramSwitchID)
	if !ok {
		return
	}
	upstreamSwitchId, ok := intParam(c, paramUpstreamSwitchID)
	if !ok {
		return
	}

	if err := ctr.Switches.DelSwitchUpstream(c.Request.Context(), userEmail, switchId, upstreamSwitchId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}
