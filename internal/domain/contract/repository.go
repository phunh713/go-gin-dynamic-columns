package contract

import (
	"context"
	"gin-demo/internal/shared/base"
)

type ContractRepository interface {
	GetById(ctx context.Context, id int64) (*Contract, error)
	GetAll(ctx context.Context) []Contract
	Create(ctx context.Context, entity *Contract) (*Contract, error)
	Update(ctx context.Context, id int64, updatePayload *ContractUpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type contractRepository struct {
	base.BaseHelper
}

func NewContractRepository() ContractRepository {
	return &contractRepository{}
}

func (r *contractRepository) GetById(ctx context.Context, id int64) (*Contract, error) {
	tx := r.GetDbTx(ctx)
	var entity Contract
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *contractRepository) GetAll(ctx context.Context) []Contract {
	tx := r.GetDbTx(ctx)
	var entities []Contract
	tx.Find(&entities)
	return entities
}

func (r *contractRepository) Create(ctx context.Context, entity *Contract) (*Contract, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *contractRepository) Update(ctx context.Context, id int64, updatePayload *ContractUpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Contract{}).Where("id = ?", id).Updates(updatePayload).Error
}

func (r *contractRepository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&Contract{}, id).Error
}
