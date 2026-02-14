package middlewares

import (
	"context"
	"gin-demo/internal/application/config"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func LogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a transaction
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
