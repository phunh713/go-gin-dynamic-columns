package middlewares

import (
	"context"
	"gin-demo/internal/application/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a transaction
		logger := config.NewLogger()
		logPayload := &config.LogPayload{}

		// Store the transaction in both gin context and request context
		c.Set(config.LogPayloadKey, logPayload)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.LogPayloadKey, logPayload)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		logger.Info("Context Log", "data", logPayload)
	}
}
