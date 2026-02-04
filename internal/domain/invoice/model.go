package invoice

import (
	"time"
)

type Invoice struct {
	Id            int64      `json:"id" gorm:"primaryKey;column:id"`
	Description   string     `json:"description" gorm:"column:description"`
	TotalAmount   float64    `json:"total_amount" gorm:"column:total_amount"`
	PendingAmount float64    `json:"pending_amount" gorm:"column:pending_amount"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	DueDate       *time.Time `json:"due_date" gorm:"column:due_date"` // nullable
	Status        string     `json:"status" gorm:"column:status"`
	PaymentTerms  int        `json:"payment_terms" gorm:"column:payment_terms" binding:"required,min=1"`
	PaidAt        *time.Time `json:"paid_at" gorm:"column:paid_at"` // nullable
	CompanyId     int64      `json:"company_id" gorm:"column:company_id" binding:"required"`
}

type InvoiceUpdateRequest struct {
	Description  *string    `json:"description,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	PaymentTerms *int       `json:"payment_terms,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CompanyId    *int64     `json:"company_id,omitempty"`
}
