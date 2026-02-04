package payment

import (
	"context"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type PaymentService interface {
	GetAllPayments(ctx context.Context) []Payment
	GetById(ctx context.Context, id int64) (*Payment, error)
	Create(ctx context.Context, entity *Payment) (*Payment, error)
}

type paymentService struct {
	paymentRepo          PaymentRepository
	dynamiccolumnService dynamiccolumn.DynamicColumnService
}

func NewPaymentService(paymentRepo PaymentRepository, dynamiccolumnService dynamiccolumn.DynamicColumnService) PaymentService {
	return &paymentService{paymentRepo: paymentRepo, dynamiccolumnService: dynamiccolumnService}
}

func (s *paymentService) GetAllPayments(ctx context.Context) []Payment {
	return s.paymentRepo.GetAll(ctx)
}

func (s *paymentService) GetById(ctx context.Context, id int64) (*Payment, error) {
	return s.paymentRepo.GetById(ctx, id)
}

func (s *paymentService) Create(ctx context.Context, entity *Payment) (*Payment, error) {
	created, err := s.paymentRepo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamiccolumnService.RefreshDynamicColumnsOfRecordId(ctx, "payment", created.Id, constants.ActionCreate, nil, nil)
	if err != nil {
		return nil, err
	}

	// Fetch updated record
	return s.paymentRepo.GetById(ctx, created.Id)
}
