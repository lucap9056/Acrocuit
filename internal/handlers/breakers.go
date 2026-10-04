package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type breakersController struct {
	controller
}

const (
	paramBreakerID = "breakerId"
	routeBreakerID = ":" + paramBreakerID
)

func breakersHandlers(r *gin.RouterGroup, base controller) {
	ctr := &breakersController{base}

	r.GET(routeBreakerID+"/downstream", ctr.GetBreakerDownstream)
	r.GET(routeBreakerID, ctr.GetBreaker)
	r.PUT(routeBreakerID, ctr.SetBreaker)
	r.DELETE(routeBreakerID, ctr.DelBreaker)
	r.PATCH(routeBreakerID+"/order", ctr.SetBreakerOrder)
	r.PATCH(routeBreakerID+"/upstream", ctr.SetBreakerUpstream)
}

type breakerDownstreamResponse struct {
	Breakers []models.Breaker `json:"breakers"`
	Devices  []models.Device  `json:"devices"`
}

func (ctr *breakersController) GetBreakerDownstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	breakers, devices, err := ctr.Breakers.GetBreakerDownstream(c.Request.Context(), userEmail, breakerId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breakerDownstreamResponse{Breakers: breakers, Devices: devices})
}

func (ctr *breakersController) GetBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}
	include := c.Query("include")

	breaker, err := ctr.Breakers.GetBreaker(c.Request.Context(), userEmail, breakerId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breaker)
}

type setBreakerRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=255"`
	DisplayOrder *int    `json:"display_order"`
}

func (ctr *breakersController) SetBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	var req setBreakerRequest
	if !bindJSON(c, &req) {
		return
	}

	breaker, err := ctr.Breakers.SetBreaker(c.Request.Context(), userEmail, breakerId, req.Name, req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breaker)
}

func (ctr *breakersController) DelBreaker(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	if err := ctr.Breakers.DelBreaker(c.Request.Context(), userEmail, breakerId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *breakersController) SetBreakerOrder(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	var req setOrderRequest
	if !bindJSON(c, &req) {
		return
	}

	breaker, err := ctr.Breakers.SetBreaker(c.Request.Context(), userEmail, breakerId, nil, &req.DisplayOrder)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breaker)
}

type setBreakerUpstreamRequest struct {
	UpstreamBreakerId *int `json:"upstream_breaker_id"`
}

func (ctr *breakersController) SetBreakerUpstream(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	breakerId, ok := intParam(c, paramBreakerID)
	if !ok {
		return
	}

	var req setBreakerUpstreamRequest
	if !bindJSON(c, &req) {
		return
	}

	breaker, err := ctr.Breakers.SetBreakerUpstream(c.Request.Context(), userEmail, breakerId, req.UpstreamBreakerId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, breaker)
}
