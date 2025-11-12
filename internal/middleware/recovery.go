package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				log.Errorw("Panic recovered",
					"error", fmt.Sprintf("%v", err),
					"stack", stack,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
