package main

import (
	"fmt"
	"gin-demo/internal/domain/dynamiccolumn"

	"gorm.io/gorm"
)

func seedInvoices(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []dynamiccolumn.DynamicColumn{
		{
			TableName: "invoices",
			Name:      "pending_amount",
			Type:      "float",
			Formula: `
			(
				SELECT COALESCE(total_amount - SUM(amount), total_amount)
				FROM payments 
				WHERE payments.invoice_id = id
			)
			`,
			Dependencies: map[string]dynamiccolumn.Dependency{
				"invoices": {
					Columns: []string{"total_amount"},
				},
				"payments": {
					Columns:           []string{"amount", "invoice_id"},
					RecordIdsSelector: "SELECT invoice_id FROM payments WHERE payments.id = {payments.id} UNION SELECT {payments:original.invoice_id} as invoice_id",
				},
			},
		},
		{
			TableName: "invoices",
			Name:      "status",
			Type:      "string",
			Formula: `
				CASE 
					WHEN pending_amount <= 0 THEN 'Paid'
					WHEN CURRENT_DATE - created_at > payment_terms * INTERVAL '1 day' THEN 'Overdue'
					ELSE 'Pending' 
				END
			`,
			Dependencies: map[string]dynamiccolumn.Dependency{
				"invoices": {
					Columns: []string{"pending_amount", "created_at", "payment_terms"},
				},
			},
		},
		{
			TableName: "invoices",
			Name:      "force_payment",
			Type:      "bool",
			Formula: `
				CASE 
					WHEN (
						SELECT status FROM companies WHERE company_id = id
					) = 'At Risk' THEN true
					ELSE false
				END
			`,
			Dependencies: map[string]dynamiccolumn.Dependency{
				"invoices": {
					Columns: []string{"id"},
				},
				"companies": {
					Columns:           []string{"status"},
					RecordIdsSelector: "SELECT id FROM invoices WHERE company_id in ({companies.ids})",
				},
			},
		},
	}

	for _, col := range dycol {
		if err := db.Create(&col).Error; err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func seedCompanies(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []dynamiccolumn.DynamicColumn{
		{
			TableName: "companies",
			Name:      "status",
			Type:      "string",
			Formula: `
				CASE
					WHEN companies.is_working = false THEN 'Inactive'
					WHEN (SELECT COUNT(*) FROM invoices WHERE invoices.company_id = companies.id AND invoices.status = 'Overdue') > 5 THEN 'At Risk'
					ELSE 'Active'
				END
			`,
			Dependencies: map[string]dynamiccolumn.Dependency{
				"companies": {
					Columns: []string{"is_working"},
				},
				"invoices": {
					Columns:           []string{"status", "company_id"},
					RecordIdsSelector: "SELECT company_id FROM invoices inv WHERE inv.id IN ({invoices.ids})",
				},
			},
		},
	}

	for _, col := range dycol {
		if err := db.Create(&col).Error; err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func SeedDynamicColumns(db *gorm.DB) {
	seedCompanies(db)
	seedInvoices(db)
}
