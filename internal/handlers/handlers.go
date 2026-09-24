package handlers

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/storage"

	"github.com/gin-gonic/gin"
)

func New(g *gin.Engine, s *storage.Storage, jwtService *auth.JWTService) {

	{
		r := g.Group("auth")
		authHandlers(r, s, jwtService)
	}

	protected := g.Group("")
	protected.Use(auth.RequireAccessToken(jwtService))

	{
		r := protected.Group("space-groups")
		spaceGroupsHandlers(r, s)
	}

	{
		r := protected.Group("spaces")
		spacesHandlers(r, s)
	}

	{
		r := protected.Group("breaker-groups")
		breakerGroupsHandlers(r, s)
	}

	{
		r := protected.Group("breakers")
		breakersHandlers(r, s)
	}

	{
		r := protected.Group("devices")
		devicesHandlers(r, s)
	}
}
