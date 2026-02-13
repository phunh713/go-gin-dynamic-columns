package deployment

import (
	"gin-demo/internal/shared/types"
	"time"
)

type Deployment struct {
	types.GormModel
	Name        string     `json:"name" gorm:"column:name" binding:"required"`
	Description string     `json:"description" gorm:"column:description"`
	ContractId  int64      `json:"contract_id" gorm:"column:contract_id" binding:"required"`
	StartDate   *time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate     *time.Time `json:"end_date" gorm:"column:end_date"`
	Status      string     `json:"status" gorm:"column:status;default:active"`
	EmployeeId  *int64     `json:"employee_id" gorm:"column:employee_id;uniqueIndex"`
	CheckinAt   *time.Time `json:"checkin_at" gorm:"column:checkin_at"`
	CheckoutAt  *time.Time `json:"checkout_at" gorm:"column:checkout_at"`
	IsCancelled bool       `json:"is_cancelled" gorm:"column:is_cancelled;default:false"`
	CanStart    bool       `json:"can_start" gorm:"column:can_start;default:false"`
}

type DeploymentUpdateRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Status      *string    `json:"status,omitempty"`
	IsCancelled *bool      `json:"is_cancelled,omitempty"`
	CheckinAt   *time.Time `json:"checkin_at,omitempty"`
	CheckoutAt  *time.Time `json:"checkout_at,omitempty"`
	EmployeeId  *int64     `json:"employee_id,omitempty"`
}
