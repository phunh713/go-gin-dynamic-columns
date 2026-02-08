package contract

import (
	"gin-demo/internal/shared/types"
	"time"
)

type Contract struct {
	types.GormModel
	Name        string    `json:"name" gorm:"column:name" binding:"required"`
	Description string    `json:"description" gorm:"column:description"`
	CompanyId   int64     `json:"company_id" gorm:"column:company_id" binding:"required"`
	Status      string    `json:"status" gorm:"column:status;default:active"` // active, completed, cancelled
	IsCancelled bool      `json:"is_cancelled" gorm:"column:is_cancelled;default:false"`
	Value       float64   `json:"value" gorm:"column:value" binding:"required"`
	StartDate   time.Time `json:"start_date" gorm:"column:start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" gorm:"column:end_date" binding:"required"`
}

type ContractUpdateRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Value       *float64   `json:"value,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}
