package deployment

import (
	"gin-demo/internal/shared/models"
	"time"
)

type Deployment struct {
	models.GormModel
	Name        string     `json:"name" gorm:"column:name" binding:"required"`
	Description string     `json:"description" gorm:"column:description"`
	ContractId  int64      `json:"contract_id" gorm:"column:contract_id" binding:"required"`
	StartDate   *time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate     *time.Time `json:"end_date" gorm:"column:end_date"`
	Status      string     `json:"status" gorm:"column:status;default:active"` // active, completed, on_hold
}

type DeploymentUpdateRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Status      *string    `json:"status,omitempty"`
}
