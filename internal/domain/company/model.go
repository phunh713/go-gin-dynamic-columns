package company

import "gin-demo/internal/shared/models"

type Company struct {
	models.GormModel
	Name     string `json:"name" gorm:"column:name" binding:"required"`
	IsActive bool   `json:"is_active" gorm:"column:is_active;default:true"`
	Status   string `json:"status" gorm:"column:status"` // Approval Pending, Active, Inactive, At Risk, Suspended (dynamic)
}

type CompanyUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
