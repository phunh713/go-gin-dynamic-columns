package base

import (
	"context"
	"gin-demo/internal/application/config"

	"gorm.io/gorm"
)

type BaseHelper struct{}

func (r *BaseHelper) GetDbTx(ctx context.Context) *gorm.DB {
	return ctx.Value(config.ContextKeyDB).(*gorm.DB)
}
