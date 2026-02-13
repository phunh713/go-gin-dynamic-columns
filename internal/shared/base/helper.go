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

func (r *BaseHelper) GetLogPayload(ctx context.Context) *config.LogPayload {
	return ctx.Value(config.LogPayloadKey).(*config.LogPayload)
}
