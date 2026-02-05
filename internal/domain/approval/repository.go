package approval

import (
	"context"
	"fmt"
	"gin-demo/internal/shared/base"
)

type ApprovalRepository interface {
	GetById(ctx context.Context, id int64) (*Approval, error)
	GetAll(ctx context.Context) []Approval
	Create(ctx context.Context, entity *Approval) (*Approval, error)
	Update(ctx context.Context, id int64, updatePayload *ApprovalUpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type approvalRepository struct {
	base.BaseHelper
}

func NewApprovalRepository() ApprovalRepository {
	return &approvalRepository{}
}

func (r *approvalRepository) GetById(ctx context.Context, id int64) (*Approval, error) {
	tx := r.GetDbTx(ctx)
	var entity Approval
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *approvalRepository) GetAll(ctx context.Context) []Approval {
	tx := r.GetDbTx(ctx)
	var entities []Approval
	tx.Find(&entities)
	return entities
}

func (r *approvalRepository) Create(ctx context.Context, entity *Approval) (*Approval, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return entity, nil
}

func (r *approvalRepository) Update(ctx context.Context, id int64, updatePayload *ApprovalUpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Approval{}).Where("id = ?", id).Updates(updatePayload).Error
}

func (r *approvalRepository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&Approval{}, id).Error
}
