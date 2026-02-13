package contract

import (
	"context"
	"fmt"
	"gin-demo/internal/shared/base"
)

type ContractRepository interface {
	GetById(ctx context.Context, id int64) (*Contract, error)
	GetAll(ctx context.Context) []Contract
	Create(ctx context.Context, entity *Contract) (*Contract, error)
	CreateMultiple(ctx context.Context, contracts []Contract) ([]Contract, error)
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
	tx.Find(&entities).Where("company_id = ?", 1)
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

func (r *contractRepository) CreateMultiple(ctx context.Context, contracts []Contract) ([]Contract, error) {
	fmt.Printf("Creating %d contracts...\n", len(contracts))
	tx := r.GetDbTx(ctx)

	// Batch insert to avoid PostgreSQL parameter limit (65535)
	// With ~9 fields per contract, we can safely do 1000 per batch
	batchSize := 5000
	createdContracts := make([]Contract, 0, len(contracts))

	for i := 0; i < len(contracts); i += batchSize {
		end := i + batchSize
		if end > len(contracts) {
			end = len(contracts)
		}

		batch := contracts[i:end]
		err := tx.Create(&batch).Error
		if err != nil {
			fmt.Printf("Error creating contract batch %d-%d: %v\n", i, end, err)
			return nil, err
		}

		createdContracts = append(createdContracts, batch...)
		fmt.Printf("Created batch %d/%d (%d contracts)\n", end, len(contracts), len(batch))
	}

	return createdContracts, nil
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
