package company

import (
	"context"
	"gin-demo/internal/shared/base"
	"gin-demo/internal/system/dynamiccolumn"
)

type CompanyRepository interface {
	GetAll(ctx context.Context) []Company
	GetById(ctx context.Context, id int64) (*Company, error)
	Create(ctx context.Context, company *Company) (*Company, error)
	Update(ctx context.Context, id int64, companyUpdate *CompanyUpdateRequest) error
}

type companyRepository struct {
	base.BaseHelper
	dynamiccolumnRepo dynamiccolumn.DynamicColumnRepository
}

func NewCompanyRepository() CompanyRepository {
	return &companyRepository{}
}

func (r *companyRepository) GetAll(ctx context.Context) []Company {
	tx := r.GetDbTx(ctx)
	var companies []Company
	tx.Find(&companies)
	return companies
}

func (r *companyRepository) Create(ctx context.Context, company *Company) (*Company, error) {
	tx := r.GetDbTx(ctx)

	err := tx.Create(company).Error
	if err != nil {
		return nil, err
	}

	return company, nil
}

func (r *companyRepository) GetById(ctx context.Context, id int64) (*Company, error) {
	tx := r.GetDbTx(ctx)
	var company Company
	err := tx.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Update(ctx context.Context, id int64, companyUpdate *CompanyUpdateRequest) error {
	tx := r.GetDbTx(ctx)

	err := tx.Model(&Company{}).Where("id = ?", id).Updates(companyUpdate).Error
	if err != nil {
		return err
	}

	return nil
}
