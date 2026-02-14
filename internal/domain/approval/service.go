package approval

import (
	"context"
	"errors"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/system/dynamiccolumn"
)

type ApprovalService interface {
	GetAll(ctx context.Context) []Approval
	GetById(ctx context.Context, id int64) (*Approval, error)
	Create(ctx context.Context, entity *Approval) (*Approval, error)
	Update(ctx context.Context, id int64, updatePayload *ApprovalUpdateRequest) (*Approval, error)
	Delete(ctx context.Context, id int64) error
}

type approvalService struct {
	approvalRepo         ApprovalRepository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func NewApprovalService(approvalRepo ApprovalRepository, dynamicColumnService dynamiccolumn.DynamicColumnService) ApprovalService {
	return &approvalService{
		approvalRepo:         approvalRepo,
		dynamicColumnService: dynamicColumnService,
	}
}

func (s *approvalService) GetAll(ctx context.Context) []Approval {
	return s.approvalRepo.GetAll(ctx)
}

func (s *approvalService) GetById(ctx context.Context, id int64) (*Approval, error) {
	return s.approvalRepo.GetById(ctx, id)
}

func (s *approvalService) Create(ctx context.Context, entity *Approval) (*Approval, error) {
	entity, err := s.approvalRepo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameApproval, []int64{entity.Id}, constants.ActionCreate, nil, entity)
	if err != nil {
		return nil, err
	}

	// Fetch updated record with dynamic columns
	refreshedEntity, err := s.approvalRepo.GetById(ctx, entity.Id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *approvalService) Update(ctx context.Context, id int64, updatePayload *ApprovalUpdateRequest) (*Approval, error) {
	originalEntity, err := s.approvalRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.approvalRepo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameApproval, []int64{id}, constants.ActionUpdate, &originalEntity.Id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Fetch updated record
	refreshedEntity, err := s.approvalRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *approvalService) Delete(ctx context.Context, id int64) error {
	originalEntity, err := s.approvalRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalEntity == nil {
		return errors.New("approval not found")
	}

	err = s.approvalRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Refresh dynamic columns after deletion
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameApproval, []int64{id}, constants.ActionDelete, &originalEntity.Id, nil)
	return err
}
