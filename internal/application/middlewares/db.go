package middlewares

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DbMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
			fmt.Println("Error starting transaction:", tx.Error)
			c.AbortWithStatusJSON(500, gin.H{"error": "database transaction error"})
			return
		}

		// Store the transaction in both gin context and request context
		c.Set(config.ContextKeyDB, tx)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.ContextKeyDB, tx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// After request is processed, commit or rollback transaction
		if c.Writer.Status() >= 400 || len(c.Errors) > 0 {
			tx.Rollback()
		} else {
			if err := tx.Commit().Error; err != nil {
				fmt.Println("Error committing transaction:", err)
				c.AbortWithStatusJSON(500, gin.H{"error": "transaction commit error"})
			}
		}
	}
}
