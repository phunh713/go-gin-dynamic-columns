package company

import (
	"context"
	"errors"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"
)

type CompanyService interface {
	GetAllCompanies(ctx context.Context) []Company
	GetById(ctx context.Context, id int64) (*Company, error)
	Create(ctx context.Context, company *Company) (*Company, error)
	Update(ctx context.Context, id int64, companyUpdate *CompanyUpdateRequest) error
}

type companyService struct {
	companyRepo          CompanyRepository
	dynamiccolumnService dynamiccolumn.DynamicColumnService
}

func NewCompanyService(companyRepo CompanyRepository, dynamiccolumnService dynamiccolumn.DynamicColumnService) CompanyService {
	return &companyService{companyRepo: companyRepo, dynamiccolumnService: dynamiccolumnService}
}

func (s *companyService) GetAllCompanies(ctx context.Context) []Company {
	return s.companyRepo.GetAll(ctx)
}

func (s *companyService) GetById(ctx context.Context, id int64) (*Company, error) {
	return s.companyRepo.GetById(ctx, id)
}

func (s *companyService) Create(ctx context.Context, company *Company) (*Company, error) {
	createdCompany, err := s.companyRepo.Create(ctx, company)
	if err != nil {
		return nil, err
	}

	// Refresh dynamic columns in database
	err = s.dynamiccolumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameCompany, []int64{createdCompany.Id}, constants.ActionCreate, nil, nil, company)
	if err != nil {
		return nil, err
	}

	// Fetch updated record with computed dynamic columns
	return s.companyRepo.GetById(ctx, createdCompany.Id)
}

func (s *companyService) Update(ctx context.Context, id int64, companyUpdate *CompanyUpdateRequest) error {
	originalCompany, err := s.companyRepo.GetById(ctx, id)
	if err != nil {
		return err
	}

	if originalCompany == nil {
		return errors.New("company not found")
	}

	err = s.companyRepo.Update(ctx, id, companyUpdate)
	if err != nil {
		return err
	}

	// Refresh dynamic columns in database
	err = s.dynamiccolumnService.RefreshDynamicColumnsOfRecordIds(ctx, constants.TableNameCompany, []int64{id}, constants.ActionUpdate, nil, &originalCompany.Id, companyUpdate)
	if err != nil {
		return err
	}

	return nil
}
