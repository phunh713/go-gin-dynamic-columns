package payment

import (
	"context"
	"gin-demo/internal/shared/base"
)

type PaymentRepository interface {
	GetById(ctx context.Context, id int64) (*Payment, error)
	GetAll(ctx context.Context) []Payment
	Create(ctx context.Context, entity *Payment) (*Payment, error)
}

type paymentRepository struct {
	base.BaseHelper
}

func NewPaymentRepository() PaymentRepository {
	return &paymentRepository{}
}

func (r *paymentRepository) GetById(ctx context.Context, id int64) (*Payment, error) {
	tx := r.GetDbTx(ctx)
	var entity Payment
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *paymentRepository) GetAll(ctx context.Context) []Payment {
	tx := r.GetDbTx(ctx)
	var entities []Payment
	tx.Find(&entities)
	return entities
}

func (r *paymentRepository) Create(ctx context.Context, entity *Payment) (*Payment, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}
