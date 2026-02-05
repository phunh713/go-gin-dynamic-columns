package payment

import (
	"context"
	"gin-demo/internal/shared/base"
)

type PaymentRepository interface {
	GetById(ctx context.Context, id int64) (*Payment, error)
	GetAll(ctx context.Context) []Payment
	Create(ctx context.Context, entity *Payment) (*Payment, error)
	Update(ctx context.Context, id int64, updatePayload *PaymentUpdateRequest) error
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
	err := tx.Raw("INSERT INTO payments (invoice_id, amount, paid_at, description) VALUES (?, ?, ?, ?) RETURNING id",
		entity.InvoiceId, entity.Amount, entity.PaidAt, entity.Description).Scan(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *paymentRepository) Update(ctx context.Context, id int64, updatePayload *PaymentUpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Payment{}).Where("id = ?", id).Updates(updatePayload).Error
}
