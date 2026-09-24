package auth

import (
	"acrocuit/internal/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const contextUserIdKey = "userId"

func RequireAccessToken(s *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || token == "" {
			response.AbortError(c, http.StatusUnauthorized, "Missing or invalid Authorization header.")
			return
		}

		userId, _, err := s.ValidateAccessToken(token)
		if err != nil {
			response.AbortError(c, http.StatusUnauthorized, "Invalid or expired access token.")
			return
		}

		c.Set(contextUserIdKey, userId)
		c.Next()
	}
}

func UserID(c *gin.Context) int {
	userId, _ := c.Get(contextUserIdKey)
	id, _ := userId.(int)
	return id
}
