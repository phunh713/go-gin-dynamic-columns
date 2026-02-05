package employee

import (
	"context"
	"errors"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type EmployeeService interface {
	GetAll(ctx context.Context) []Employee
	GetById(ctx context.Context, id int64) (*Employee, error)
	Create(ctx context.Context, entity *Employee) (*Employee, error)
	Update(ctx context.Context, id int64, updatePayload *EmployeeUpdateRequest) (*Employee, error)
	Delete(ctx context.Context, id int64) error
}

type employeeService struct {
	employeeRepo      EmployeeRepository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func NewEmployeeService(employeeRepo EmployeeRepository, dynamicColumnService dynamiccolumn.DynamicColumnService) EmployeeService {
	return &employeeService{
		employeeRepo:      employeeRepo,
		dynamicColumnService: dynamicColumnService,
	}
}

func (s *employeeService) GetAll(ctx context.Context) []Employee {
	return s.employeeRepo.GetAll(ctx)
}

func (s *employeeService) GetById(ctx context.Context, id int64) (*Employee, error) {
	return s.employeeRepo.GetById(ctx, id)
}

func (s *employeeService) Create(ctx context.Context, entity *Employee) (*Employee, error) {
	entity, err := s.employeeRepo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	
	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "employees", []int64{entity.Id}, constants.ActionCreate, nil, nil, entity)
	if err != nil {
		return nil, err
	}
	
	// Fetch updated record with dynamic columns
	refreshedEntity, err := s.employeeRepo.GetById(ctx, entity.Id)
	if err != nil {
		return nil, err
	}
	
	return refreshedEntity, nil
}

func (s *employeeService) Update(ctx context.Context, id int64, updatePayload *EmployeeUpdateRequest) (*Employee, error) {
	originalEntity, err := s.employeeRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	
	err = s.employeeRepo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}
	
	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "employees", []int64{id}, constants.ActionUpdate, nil, &originalEntity.Id, updatePayload)
	if err != nil {
		return nil, err
	}
	
	// Fetch updated record
	refreshedEntity, err := s.employeeRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return refreshedEntity, nil
}

func (s *employeeService) Delete(ctx context.Context, id int64) error {
	originalEntity, err := s.employeeRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalEntity == nil {
		return errors.New("employee not found")
	}
	
	err = s.employeeRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	
	// Refresh dynamic columns after deletion
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "employees", []int64{id}, constants.ActionDelete, nil, &originalEntity.Id, nil)
	return err
}
