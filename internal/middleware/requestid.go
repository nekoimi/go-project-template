package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
)

const RequestIDKey = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = idgen.GenerateStringID()
		}
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDKey, requestID)
		c.Next()
	}
}
