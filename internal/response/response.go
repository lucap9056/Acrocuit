package response

import (
	"acrocuit/internal/logs"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		logs.Out.Error("panic recovered", zap.Any("panic", recovered), zap.String("path", c.Request.URL.Path), zap.Stack("stack"))
		AbortError(c, http.StatusInternalServerError, "internal server error")
	})
}
