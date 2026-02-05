package approval

import (
	"gin-demo/internal/shared/models"
	"time"
)

type Approval struct {
	models.GormModel
	CompanyId    int64      `json:"company_id" gorm:"column:company_id" binding:"required"`
	ApproverName string     `json:"approver_name" gorm:"column:approver_name" binding:"required"`
	Status       string     `json:"status" gorm:"column:status;default:pending"` // pending, approved, rejected
	Comments     string     `json:"comments" gorm:"column:comments"`
	ReviewedAt   *time.Time `json:"reviewed_at" gorm:"column:reviewed_at"`
}

type ApprovalUpdateRequest struct {
	ApproverName *string    `json:"approver_name,omitempty"`
	Status       *string    `json:"status,omitempty"`
	Comments     *string    `json:"comments,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}
