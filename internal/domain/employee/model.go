package employee

import "gin-demo/internal/shared/models"

type Employee struct {
	models.GormModel
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

// Join table for many-to-many relationship
type EmployeeDeployment struct {
	Id           int64  `json:"id" gorm:"primaryKey;column:id"`
	EmployeeId   int64  `json:"employee_id" gorm:"column:employee_id;uniqueIndex:idx_employee_deployment"`
	DeploymentId int64  `json:"deployment_id" gorm:"column:deployment_id;uniqueIndex:idx_employee_deployment"`
	Role         string `json:"role" gorm:"column:role"` // developer, manager, consultant, etc.
	CreatedAt    int64  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}
