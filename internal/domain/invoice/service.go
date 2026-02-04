package invoice

import (
	"context"
	"errors"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type InvoiceService interface {
	GetAll(ctx context.Context) []Invoice
	GetById(ctx context.Context, id int64) (*Invoice, error)
	Create(ctx context.Context, invoice *Invoice) (*Invoice, error)
	Update(ctx context.Context, id int64, updatePayload *InvoiceUpdateRequest) (*Invoice, error)
	Delete(ctx context.Context, id int64) error
}

type invoiceService struct {
	invoiceRepo          InvoiceRepository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func NewInvoiceService(invoiceRepo InvoiceRepository, dynamicColumnService dynamiccolumn.DynamicColumnService) InvoiceService {
	return &invoiceService{invoiceRepo: invoiceRepo, dynamicColumnService: dynamicColumnService}
}

func (s *invoiceService) GetAll(ctx context.Context) []Invoice {
	return s.invoiceRepo.GetAll(ctx)
}

func (s *invoiceService) GetById(ctx context.Context, id int64) (*Invoice, error) {
	return s.invoiceRepo.GetById(ctx, id)
}

func (s *invoiceService) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	invoice, err := s.invoiceRepo.Create(ctx, invoice)
	if err != nil {
		return nil, err
	}
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordId(ctx, "invoices", invoice.Id, constants.ActionCreate, nil, nil)
	if err != nil {
		return nil, err
	}
	refreshedInvoice, err := s.invoiceRepo.GetById(ctx, invoice.Id)
	if err != nil {
		return nil, err
	}
	return refreshedInvoice, nil
}

func (s *invoiceService) Update(ctx context.Context, id int64, updatePayload *InvoiceUpdateRequest) (*Invoice, error) {
	originalInvoice, err := s.invoiceRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	err = s.invoiceRepo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordId(ctx, "invoices", id, constants.ActionUpdate, nil, originalInvoice)
	if err != nil {
		return nil, err
	}
	refreshedInvoice, err := s.invoiceRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return refreshedInvoice, nil
}

func (s *invoiceService) Delete(ctx context.Context, id int64) error {
	originalInvoice, err := s.invoiceRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalInvoice == nil {
		return errors.New("invoice not found")
	}
	err = s.invoiceRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordId(ctx, "invoices", id, constants.ActionDelete, nil, originalInvoice)
	return err
}
