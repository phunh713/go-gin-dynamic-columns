package payment

import (
	"gin-demo/internal/shared/models"
	"time"
)

type Payment struct {
	models.GormModel
	Description string    `json:"description" gorm:"column:description"`
	Amount      float64   `json:"amount" gorm:"column:amount" binding:"required"`
	PaidAt      time.Time `json:"paid_at" gorm:"column:paid_at"`
	InvoiceId   int64     `json:"invoice_id" gorm:"column:invoice_id" binding:"required"`
}

type PaymentUpdateRequest struct {
	Description *string    `json:"description,omitempty"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	Amount      *float64   `json:"amount,omitempty"`
	InvoiceId   *int64     `json:"invoice_id,omitempty"`
}
