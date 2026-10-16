package handlers

import (
	"acrocuit/internal/response"

	"github.com/gin-gonic/gin"
)

type metaController struct {
	controller
}

type metaResponse struct {
	AuthRequired bool `json:"auth_required"`
}

func metaHandlers(r *gin.RouterGroup, base controller) {
	ctr := &metaController{base}

	r.GET("", ctr.GetMeta)
}

func (ctr *metaController) GetMeta(c *gin.Context) {
	response.OK(c, metaResponse{AuthRequired: ctr.options.IsMultiUser()})
}
