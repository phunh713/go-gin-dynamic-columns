package invoice

import (
	"context"
	"errors"
	"gin-demo/internal/shared/base"
)

type InvoiceRepository interface {
	GetById(ctx context.Context, id int64) (*Invoice, error)
	GetAll(ctx context.Context) []Invoice
	Create(ctx context.Context, entity *Invoice) (*Invoice, error)
	Update(ctx context.Context, id int64, invoice *InvoiceUpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type invoiceRepository struct {
	base.BaseHelper
}

func NewInvoiceRepository() InvoiceRepository {
	return &invoiceRepository{}
}

func (r *invoiceRepository) GetById(ctx context.Context, id int64) (*Invoice, error) {
	tx := r.GetDbTx(ctx)
	var invoice Invoice
	err := tx.First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) GetAll(ctx context.Context) []Invoice {
	tx := r.GetDbTx(ctx)
	var invoices []Invoice
	tx.Find(&invoices)
	return invoices
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	tx := r.GetDbTx(ctx).Debug()
	err := tx.Create(invoice).Error

	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (r *invoiceRepository) Update(ctx context.Context, id int64, invoiceUpdate *InvoiceUpdateRequest) error {

	if id <= 0 {
		return errors.New("invalid invoice id")
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Invoice{}).Where("id = ?", id).Updates(invoiceUpdate).Error
}

func (r *invoiceRepository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&Invoice{}, id).Error
}
