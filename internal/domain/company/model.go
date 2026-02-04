package company

import "time"

type Company struct {
	Id        int64     `json:"id" gorm:"primaryKey;column:id"`
	Name      string    `json:"name" gorm:"column:name"`
	IsWorking bool      `json:"is_working" gorm:"column:is_working"`
	Status    string    `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

type CompanyUpdateRequest struct {
	Name      *string `json:"name,omitempty" gorm:"column:name"`
	IsWorking *bool   `json:"is_working,omitempty" gorm:"column:is_working"`
}
