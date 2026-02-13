package deployment

import (
	"context"
	"errors"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type DeploymentService interface {
	GetAll(ctx context.Context) []Deployment
	GetById(ctx context.Context, id int64) (*Deployment, error)
	Create(ctx context.Context, entity *Deployment) (*Deployment, error)
	Update(ctx context.Context, id int64, updatePayload *DeploymentUpdateRequest) (*Deployment, error)
	Delete(ctx context.Context, id int64) error
}

type deploymentService struct {
	deploymentRepo       DeploymentRepository
	dynamicColumnService dynamiccolumn.DynamicColumnService
}

func NewDeploymentService(deploymentRepo DeploymentRepository, dynamicColumnService dynamiccolumn.DynamicColumnService) DeploymentService {
	return &deploymentService{
		deploymentRepo:       deploymentRepo,
		dynamicColumnService: dynamicColumnService,
	}
}

func (s *deploymentService) GetAll(ctx context.Context) []Deployment {
	return s.deploymentRepo.GetAll(ctx)
}

func (s *deploymentService) GetById(ctx context.Context, id int64) (*Deployment, error) {
	return s.deploymentRepo.GetById(ctx, id)
}

func (s *deploymentService) Create(ctx context.Context, entity *Deployment) (*Deployment, error) {
	entity, err := s.deploymentRepo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameDeployment, []int64{entity.Id}, constants.ActionCreate, nil, entity)
	if err != nil {
		return nil, err
	}

	// Fetch updated record with dynamic columns
	refreshedEntity, err := s.deploymentRepo.GetById(ctx, entity.Id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *deploymentService) Update(ctx context.Context, id int64, updatePayload *DeploymentUpdateRequest) (*Deployment, error) {
	originalEntity, err := s.deploymentRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.deploymentRepo.Update(ctx, id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameDeployment, []int64{id}, constants.ActionUpdate, &originalEntity.Id, updatePayload)
	if err != nil {
		return nil, err
	}

	// Fetch updated record
	refreshedEntity, err := s.deploymentRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return refreshedEntity, nil
}

func (s *deploymentService) Delete(ctx context.Context, id int64) error {
	originalEntity, err := s.deploymentRepo.GetById(ctx, id)
	if err != nil {
		return err
	}
	if originalEntity == nil {
		return errors.New("deployment not found")
	}

	err = s.deploymentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Refresh dynamic columns after deletion
	err = s.dynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameDeployment, []int64{id}, constants.ActionDelete, &originalEntity.Id, nil)
	return err
}
