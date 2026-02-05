package invoice

import (
	"gin-demo/internal/shared/models"
	"time"
)

type Invoice struct {
	models.GormModel
	InvoiceNumber string     `json:"invoice_number" gorm:"column:invoice_number;uniqueIndex"`
	Description   string     `json:"description" gorm:"column:description"`
	TotalAmount   float64    `json:"total_amount" gorm:"column:total_amount"`
	PendingAmount float64    `json:"pending_amount" gorm:"column:pending_amount"` // dynamic column
	DueDate       *time.Time `json:"due_date" gorm:"column:due_date"`
	Status        string     `json:"status" gorm:"column:status"` // Pending, Paid, Overdue (dynamic)
	PaymentTerms  int        `json:"payment_terms" gorm:"column:payment_terms;default:30" binding:"min=1"`
	PaidAt        *time.Time `json:"paid_at" gorm:"column:paid_at"`
	ContractId    int64      `json:"contract_id" gorm:"column:contract_id" binding:"required"`
}

type InvoiceUpdateRequest struct {
	InvoiceNumber *string    `json:"invoice_number,omitempty"`
	Description   *string    `json:"description,omitempty"`
	TotalAmount   *float64   `json:"total_amount,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	PaymentTerms  *int       `json:"payment_terms,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}
