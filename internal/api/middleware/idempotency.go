package middleware

import (
	"github.com/gin-gonic/gin"
)

func IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		idempotencyKey := c.GetHeader("Idempotency-Key")
		c.Set("idempotency_key", idempotencyKey)
		c.Next()
	}
}
