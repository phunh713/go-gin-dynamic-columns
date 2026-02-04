package payment

import "time"

type Payment struct {
	Id          int64     `json:"id" gorm:"primaryKey;column:id"`
	Description string    `json:"description" gorm:"column:description"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	PaidAt      time.Time `json:"paid_at" gorm:"column:paid_at"`
	Amount      float64   `json:"amount" gorm:"column:amount"`
	InvoiceId   int64     `json:"invoice_id" gorm:"column:invoice_id"`
}
