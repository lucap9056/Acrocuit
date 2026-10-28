package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	paramMemberEmail = "memberEmail"
	routeMemberEmail = ":" + paramMemberEmail
	memberEmailSelf  = "me"
)

func spaceGroupACLsHandlers(r *gin.RouterGroup, ctr *spaceGroupsController) {
	r.GET(routeSpaceGroupID+"/members", ctr.GetSpaceGroupMembers)
	r.POST(routeSpaceGroupID+"/members", ctr.AddSpaceGroupMember)
	r.PUT(routeSpaceGroupID+"/members/"+routeMemberEmail, ctr.SetSpaceGroupMemberRole)
	r.DELETE(routeSpaceGroupID+"/members/"+routeMemberEmail, ctr.DelSpaceGroupMember)
	r.PUT(routeSpaceGroupID+"/owner", ctr.TransferSpaceGroupOwner)
	r.PUT(routeSpaceGroupID+"/public", ctr.SetSpaceGroupPublic)
}

type addSpaceGroupMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=viewer editor"`
}

type spaceGroupMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=viewer editor"`
}

type transferSpaceGroupOwnerRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type spaceGroupPublicRequest struct {
	Public *bool `json:"public" binding:"required"`
}

func (ctr *spaceGroupsController) GetSpaceGroupMembers(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	members, err := ctr.SpaceGroups.GetSpaceGroupMembers(c.Request.Context(), userEmail, spaceGroupId)
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.OK(c, members)
}

func (ctr *spaceGroupsController) AddSpaceGroupMember(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	var req addSpaceGroupMemberRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.SpaceGroups.AddSpaceGroupMember(c.Request.Context(), userEmail, spaceGroupId, req.Email, req.Role); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *spaceGroupsController) SetSpaceGroupMemberRole(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}
	memberEmail := c.Param(paramMemberEmail)

	var req spaceGroupMemberRoleRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.SpaceGroups.SetSpaceGroupMemberRole(c.Request.Context(), userEmail, spaceGroupId, memberEmail, req.Role); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *spaceGroupsController) DelSpaceGroupMember(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}
	memberEmail := c.Param(paramMemberEmail)

	var err error
	if memberEmail == memberEmailSelf || memberEmail == userEmail {
		err = ctr.SpaceGroups.LeaveSpaceGroup(c.Request.Context(), userEmail, spaceGroupId)
	} else {
		err = ctr.SpaceGroups.DelSpaceGroupMember(c.Request.Context(), userEmail, spaceGroupId, memberEmail)
	}
	if err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *spaceGroupsController) TransferSpaceGroupOwner(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	var req transferSpaceGroupOwnerRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.SpaceGroups.TransferSpaceGroupOwner(c.Request.Context(), userEmail, spaceGroupId, req.Email); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}

func (ctr *spaceGroupsController) SetSpaceGroupPublic(c *gin.Context) {
	userEmail := auth.UserEmail(c)
	spaceGroupId, ok := intParam(c, paramSpaceGroupID)
	if !ok {
		return
	}

	var req spaceGroupPublicRequest
	if !bindJSON(c, &req) {
		return
	}

	if err := ctr.SpaceGroups.SetSpaceGroupPublic(c.Request.Context(), userEmail, spaceGroupId, *req.Public); err != nil {
		handleStorageErr(c, err)
		return
	}

	response.Empty(c, http.StatusOK)
}
