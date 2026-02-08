package employee

import "gin-demo/internal/shared/types"

type Employee struct {
	types.GormModel
	Name     string `json:"name" gorm:"column:name" binding:"required"`
	Email    string `json:"email" gorm:"column:email;uniqueIndex"`
	Position string `json:"position" gorm:"column:position"`
	Status   string `json:"status" gorm:"column:status;default:active"` // active, inactive
}

type EmployeeUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Position *string `json:"position,omitempty"`
	Status   *string `json:"status,omitempty"`
}
