package dynamiccolumn

import (
	"fmt"
	"gin-demo/internal/domain/dynamiccolumn"
	domainDynamicColumn "gin-demo/internal/domain/dynamiccolumn"

	"gorm.io/gorm"
)

func seedInvoices(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "invoices",
			Name:      "pending_amount",
			Type:      "float",
			Formula: `
			(
				SELECT COALESCE(invoices.total_amount - SUM(p.amount), invoices.total_amount)
				FROM payments p
				WHERE p.invoice_id = invoices.id
			)
			`,
			Dependencies: map[string]dynamiccolumn.Dependency{
				"invoices": {
					Columns: []string{"total_amount"},
				},
				"payments": {
					Columns:           []string{"amount", "invoice_id"},
					RecordIdsSelector: "SELECT invoice_id FROM payments WHERE payments.id IN ({payments.ids}) UNION SELECT {payments:original.invoice_id} as invoice_id",
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
			Dependencies: map[string]domainDynamicColumn.Dependency{
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
						SELECT status FROM companies c WHERE c.id = company_id
					) = 'At Risk' THEN true
					ELSE false
				END
			`,
			Dependencies: map[string]domainDynamicColumn.Dependency{
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
