package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/models"
	"acrocuit/internal/response"
	"acrocuit/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type spaceGroupsController struct {
	*storage.Storage
}

const (
	paramSpaceGroupID = "spaceGroupId"
	routeSpaceGroupID = ":" + paramSpaceGroupID
)

func spaceGroupsHandlers(r *gin.RouterGroup, s *storage.Storage) {
	ctr := &spaceGroupsController{s}

	r.GET("", ctr.GetSpaceGroups)
	r.POST("", ctr.AddSpaceGroup)
	r.GET(routeSpaceGroupID, ctr.GetSpaceGroup)
	r.PUT(routeSpaceGroupID, ctr.EditSpaceGroupName)
	r.DELETE(routeSpaceGroupID, ctr.DelSpaceGroup)

	r.GET(routeSpaceGroupID+"/spaces", ctr.GetSpaces)
	r.POST(routeSpaceGroupID+"/spaces", ctr.AddSpace)
}

type spaceGroupNameRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (ctr *spaceGroupsController) GetSpaceGroups(c *gin.Context) {
	userId := auth.UserID(c)
	include := c.Query("include")

	spaceGroups, err := ctr.SpaceGroups.GetSpaceGroups(c.Request.Context(), userId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, spaceGroups)
}

func (ctr *spaceGroupsController) AddSpaceGroup(c *gin.Context) {
	userId := auth.UserID(c)

	var req spaceGroupNameRequest
	if !bindJSON(c, &req) {
		return
	}

	id, err := ctr.SpaceGroups.AddSpaceGroup(c.Request.Context(), userId, req.Name)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, models.SpaceGroup{Id: id, Name: req.Name})
}

func (ctr *spaceGroupsController) GetSpaceGroup(c *gin.Context) {
	userId := auth.UserID(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}
	include := c.Query("include")

	spaceGroup, err := ctr.SpaceGroups.GetSpaceGroup(c.Request.Context(), userId, spaceGroupId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, spaceGroup)
}

func (ctr *spaceGroupsController) EditSpaceGroupName(c *gin.Context) {
	userId := auth.UserID(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	var req spaceGroupNameRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.SpaceGroups.EditSpaceGroupName(c.Request.Context(), userId, spaceGroupId, req.Name); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, models.SpaceGroup{Id: spaceGroupId, Name: req.Name})
}

func (ctr *spaceGroupsController) DelSpaceGroup(c *gin.Context) {
	userId := auth.UserID(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	if err := ctr.SpaceGroups.DelSpaceGroup(c.Request.Context(), userId, spaceGroupId); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *spaceGroupsController) GetSpaces(c *gin.Context) {
	userId := auth.UserID(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	if _, err := ctr.SpaceGroups.GetSpaceGroup(c.Request.Context(), userId, spaceGroupId, "id"); err != nil {
		handleStorageErr(c, err)
		return
	}

	include := c.Query("include")
	spaces, err := ctr.Spaces.GetSpaces(c.Request.Context(), userId, spaceGroupId, include)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, spaces)
}

type createSpaceRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (ctr *spaceGroupsController) AddSpace(c *gin.Context) {
	userId := auth.UserID(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	var req createSpaceRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := ctr.Spaces.AddSpace(c.Request.Context(), userId, spaceGroupId, req.Name)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, space)
}
