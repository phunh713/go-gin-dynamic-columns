package invoice

import (
	"context"
	"errors"
	"fmt"
	"gin-demo/internal/shared/base"
)

type InvoiceRepository interface {
	GetById(ctx context.Context, id int64) (*Invoice, error)
	GetAll(ctx context.Context) []Invoice
	Create(ctx context.Context, entity *Invoice) (*Invoice, error)
	Update(ctx context.Context, id int64, invoice *InvoiceUpdateRequest) error
	Delete(ctx context.Context, id int64) error
	CreateMultiple(ctx context.Context, invoices []Invoice) ([]Invoice, error)
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
	tx := r.GetDbTx(ctx)
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

func (r *invoiceRepository) CreateMultiple(ctx context.Context, invoices []Invoice) ([]Invoice, error) {
	fmt.Printf("Creating %d invoices...\n", len(invoices))
	tx := r.GetDbTx(ctx)

	// Batch insert to avoid PostgreSQL parameter limit (65535)
	// With ~9 fields per invoice, we can safely do 1000 per batch
	batchSize := 5000
	createdInvoices := make([]Invoice, 0, len(invoices))

	for i := 0; i < len(invoices); i += batchSize {
		end := i + batchSize
		if end > len(invoices) {
			end = len(invoices)
		}

		batch := invoices[i:end]
		err := tx.Create(&batch).Error
		if err != nil {
			fmt.Printf("Error creating invoice batch %d-%d: %v\n", i, end, err)
			return nil, err
		}

		createdInvoices = append(createdInvoices, batch...)
		fmt.Printf("Created batch %d/%d (%d invoices)\n", end, len(invoices), len(batch))
	}

	return createdInvoices, nil
}
