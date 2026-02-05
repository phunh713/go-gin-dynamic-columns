package contract

import "gin-demo/internal/shared/models"

type Contract struct {
	models.GormModel
	Name        string `json:"name" gorm:"column:name" binding:"required"`
	Description string `json:"description" gorm:"column:description"`
	CompanyId   int64  `json:"company_id" gorm:"column:company_id" binding:"required"`
	Status      string `json:"status" gorm:"column:status;default:active"` // active, completed, cancelled
}

type ContractUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}
