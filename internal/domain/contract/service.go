package contract

import (
	"context"
	"errors"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/system/dynamiccolumn"
)

type ContractService interface {
	GetAll(ctx context.Context) []Contract
	GetById(ctx context.Context, id int64) (*Contract, error)
	Create(ctx context.Context, entity *Contract) (*Contract, error)
	CreateMultiple(ctx context.Context, contracts []Contract) ([]Contract, error)
	Update(ctx context.Context, id int64, updatePayload *ContractUpdateRequest) (*Contract, error)
	Delete(ctx context.Context, id int64) error
}

type contractService struct {
	contractRepo         ContractRepository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func NewContractService(contractRepo ContractRepository, dynamicColumnService dynamiccolumn.DynamicColumnService) ContractService {
	return &contractService{
		contractRepo:         contractRepo,
		dynamicColumnService: dynamicColumnService,
	}
}

func (s *contractService) GetAll(ctx context.Context) []Contract {
	return s.contractRepo.GetAll(ctx)
}

func (s *contractService) GetById(ctx context.Context, id int64) (*Contract, error) {
	return s.contractRepo.GetById(ctx, id)
}

func (s *contractService) Create(ctx context.Context, entity *Contract) (*Contract, error) {
	entity, err := s.contractRepo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameContract, []int64{entity.Id}, constants.ActionCreate, nil, entity)
	if err != nil {
		return nil, err
	}

	// Fetch updated record with dynamic columns
	refreshedEntity, err := s.contractRepo.GetById(ctx, entity.Id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *contractService) Update(ctx context.Context, id int64, updatePayload *ContractUpdateRequest) (*Contract, error) {
	originalEntity, err := s.contractRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.contractRepo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameContract, []int64{id}, constants.ActionUpdate, &originalEntity.Id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Fetch updated record
	refreshedEntity, err := s.contractRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *contractService) Delete(ctx context.Context, id int64) error {
	originalEntity, err := s.contractRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalEntity == nil {
		return errors.New("contract not found")
	}

	err = s.contractRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Refresh dynamic columns after deletion
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameContract, []int64{id}, constants.ActionDelete, &originalEntity.Id, nil)
	return err
}

func (s *contractService) CreateMultiple(ctx context.Context, contracts []Contract) ([]Contract, error) {
	createdContracts, err := s.contractRepo.CreateMultiple(ctx, contracts)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, contract := range createdContracts {
		ids = append(ids, contract.Id)
	}
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameContract, ids, constants.ActionCreate, nil, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
