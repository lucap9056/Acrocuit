package handlers

import (
	"acrocuit/internal/models"
	"acrocuit/internal/response"

	"github.com/gin-gonic/gin"
)

type metaController struct {
	controller
}

func metaHandlers(r *gin.RouterGroup, base controller) {
	ctr := &metaController{base}

	r.GET("", ctr.GetMeta)
}

func (ctr *metaController) GetMeta(c *gin.Context) {
	response.OK(c, models.Meta{AuthRequired: ctr.options.IsMultiUser()})
}
