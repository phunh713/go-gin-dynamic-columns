package employee

import (
	"context"
	"gin-demo/internal/shared/base"
)

type EmployeeRepository interface {
	GetById(ctx context.Context, id int64) (*Employee, error)
	GetAll(ctx context.Context) []Employee
	Create(ctx context.Context, entity *Employee) (*Employee, error)
	Update(ctx context.Context, id int64, updatePayload *EmployeeUpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type employeeRepository struct {
	base.BaseHelper
}

func NewEmployeeRepository() EmployeeRepository {
	return &employeeRepository{}
}

func (r *employeeRepository) GetById(ctx context.Context, id int64) (*Employee, error) {
	tx := r.GetDbTx(ctx)
	var entity Employee
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *employeeRepository) GetAll(ctx context.Context) []Employee {
	tx := r.GetDbTx(ctx)
	var entities []Employee
	tx.Find(&entities)
	return entities
}

func (r *employeeRepository) Create(ctx context.Context, entity *Employee) (*Employee, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *employeeRepository) Update(ctx context.Context, id int64, updatePayload *EmployeeUpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Employee{}).Where("id = ?", id).Updates(updatePayload).Error
}

func (r *employeeRepository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&Employee{}, id).Error
}
