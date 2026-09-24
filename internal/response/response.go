package response

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error,omitempty"`
}

func OK[T any](c *gin.Context, data T) {
	Status(c, 200, data)
}

func Status[T any](c *gin.Context, status int, data T) {
	c.JSON(status, Response[T]{Success: true, Data: data})
}

func Empty(c *gin.Context, status int) {
	c.JSON(status, Response[any]{Success: true})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Response[any]{Success: false, Error: message})
}

func AbortError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Response[any]{Success: false, Error: message})
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v\n%s", recovered, debug.Stack())
		AbortError(c, http.StatusInternalServerError, "internal server error")
	})
}
