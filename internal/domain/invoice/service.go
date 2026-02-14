package invoice

import (
	"context"
	"errors"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/system/dynamiccolumn"
)

type InvoiceService interface {
	GetAll(ctx context.Context) []Invoice
	GetById(ctx context.Context, id int64) (*Invoice, error)
	Create(ctx context.Context, invoice *Invoice) (*Invoice, error)
	Update(ctx context.Context, id int64, updatePayload *InvoiceUpdateRequest) (*Invoice, error)
	Delete(ctx context.Context, id int64) error
	CreateMultiple(ctx context.Context, invoices []Invoice) ([]Invoice, error)
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
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameInvoice, []int64{invoice.Id}, constants.ActionCreate, nil, invoice)
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
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameInvoice, []int64{id}, constants.ActionUpdate, &originalInvoice.Id, updatePayload)
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
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameInvoice, []int64{id}, constants.ActionDelete, &originalInvoice.Id, nil)
	return err
}

func (s *invoiceService) CreateMultiple(ctx context.Context, invoices []Invoice) ([]Invoice, error) {
	createdInvoices, err := s.invoiceRepo.CreateMultiple(ctx, invoices)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, invoice := range createdInvoices {
		ids = append(ids, invoice.Id)
	}
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameInvoice, ids, constants.ActionCreate, nil, nil)
	if err != nil {
		return nil, err
	}
	refreshedInvoices := s.invoiceRepo.GetAll(ctx)

	return refreshedInvoices, nil
}
