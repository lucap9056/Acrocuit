package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type breakerGroupsController struct {
	controller
}

const (
	paramBreakerGroupID = "breakerGroupId"
	routeBreakerGroupID = ":" + paramBreakerGroupID
)

func breakerGroupsHandlers(r *gin.RouterGroup, base controller) {
	ctr := &breakerGroupsController{base}

	r.GET(routeBreakerGroupID, ctr.GetBreakerGroup)
	r.PUT(routeBreakerGroupID, ctr.SetBreakerGroup)
	r.PUT(routeBreakerGroupID+"/position", ctr.SetBreakerGroupPosition)
	r.DELETE(routeBreakerGroupID, ctr.DelBreakerGroup)

	r.GET(routeBreakerGroupID+"/breakers", ctr.GetBreakers)
	r.POST(routeBreakerGroupID+"/breakers", ctr.AddBreaker)
}

func (ctr *breakerGroupsController) GetBreakerGroup(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}
	include := c.Query("include")

	breakerGroup, err := ctr.BreakerGroups.GetBreakerGroup(c.Request.Context(), userEmail, breakerGroupId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerGroup)
}

type setBreakerGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (ctr *breakerGroupsController) SetBreakerGroup(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}

	var req setBreakerGroupRequest
	if !bindJSON(c, &req) {
		return
	}

	breakerGroup, err := ctr.BreakerGroups.SetBreakerGroup(c.Request.Context(), userEmail, breakerGroupId, req.Name)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerGroup)
}

type setPositionRequest struct {
	Position *models.Position `json:"position" binding:"required"`
}

func (ctr *breakerGroupsController) SetBreakerGroupPosition(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}

	var req setPositionRequest
	if !bindJSON(c, &req) {
		return
	}

	breakerGroup, err := ctr.BreakerGroups.SetBreakerGroupPosition(c.Request.Context(), userEmail, breakerGroupId, *req.Position)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerGroup)
}

func (ctr *breakerGroupsController) DelBreakerGroup(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}

	if err := ctr.BreakerGroups.DelBreakerGroup(c.Request.Context(), userEmail, breakerGroupId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *breakerGroupsController) GetBreakers(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}
	include := c.Query("include")

	breakers, err := ctr.Breakers.GetBreakers(c.Request.Context(), userEmail, breakerGroupId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakers)
}

type createBreakerRequest struct {
	Name              string `json:"name" binding:"required,min=1,max=255"`
	UpstreamBreakerId int    `json:"upstream_breaker_id"`
}

func (ctr *breakerGroupsController) AddBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerGroupId, ok := intParam(c, paramBreakerGroupID)
	if !ok {
		return
	}

	var req createBreakerRequest
	if !bindJSON(c, &req) {
		return
	}

	var upstreamBreakerId *int
	if req.UpstreamBreakerId != 0 {
		upstreamBreakerId = &req.UpstreamBreakerId
	}

	ctx := c.Request.Context()
	breaker, err := ctr.Breakers.AddBreaker(ctx, userEmail, breakerGroupId, req.Name, upstreamBreakerId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breaker)
}
