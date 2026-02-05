package deployment

import (
	"context"
	"gin-demo/internal/shared/base"
)

type DeploymentRepository interface {
	GetById(ctx context.Context, id int64) (*Deployment, error)
	GetAll(ctx context.Context) []Deployment
	Create(ctx context.Context, entity *Deployment) (*Deployment, error)
	Update(ctx context.Context, id int64, updatePayload *DeploymentUpdateRequest) error
	Delete(ctx context.Context, id int64) error
}

type deploymentRepository struct {
	base.BaseHelper
}

func NewDeploymentRepository() DeploymentRepository {
	return &deploymentRepository{}
}

func (r *deploymentRepository) GetById(ctx context.Context, id int64) (*Deployment, error) {
	tx := r.GetDbTx(ctx)
	var entity Deployment
	err := tx.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *deploymentRepository) GetAll(ctx context.Context) []Deployment {
	tx := r.GetDbTx(ctx)
	var entities []Deployment
	tx.Find(&entities)
	return entities
}

func (r *deploymentRepository) Create(ctx context.Context, entity *Deployment) (*Deployment, error) {
	tx := r.GetDbTx(ctx)
	err := tx.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *deploymentRepository) Update(ctx context.Context, id int64, updatePayload *DeploymentUpdateRequest) error {
	if id <= 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	return tx.Model(&Deployment{}).Where("id = ?", id).Updates(updatePayload).Error
}

func (r *deploymentRepository) Delete(ctx context.Context, id int64) error {
	tx := r.GetDbTx(ctx)
	return tx.Delete(&Deployment{}, id).Error
}
