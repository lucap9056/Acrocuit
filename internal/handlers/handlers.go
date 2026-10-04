package handlers

import (
	"acrocuit/internal/options"
	"acrocuit/internal/storage"

	"github.com/gin-gonic/gin"
)

type controller struct {
	*storage.Storage
	options *options.Options
}

func New(g *gin.Engine, s *storage.Storage, requireIdentity gin.HandlerFunc, opts *options.Options) {
	base := controller{s, opts}

	protected := g.Group("api")
	protected.Use(requireIdentity)

	{
		r := protected.Group("space-groups")
		spaceGroupsHandlers(r, base)
	}

	{
		r := protected.Group("spaces")
		spacesHandlers(r, base)
	}

	{
		r := protected.Group("breaker-groups")
		breakerGroupsHandlers(r, base)
	}

	{
		r := protected.Group("breakers")
		breakersHandlers(r, base)
	}

	{
		r := protected.Group("devices")
		devicesHandlers(r, base)
	}
}
