package dynamiccolumn

import (
	"fmt"
	domainDynamicColumn "gin-demo/internal/domain/dynamiccolumn"

	"gorm.io/gorm"
)

func seedCompanies(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "companies",
			Name:      "status",
			Type:      "string",
			Formula: `
				CASE
					WHEN companies.is_active = false THEN 'Inactive'
					WHEN (SELECT COUNT(*) FROM approvals WHERE approvals.company_id = companies.id) = 0 THEN 'No Approval'
					WHEN (SELECT COUNT(*) FROM approvals WHERE approvals.company_id = companies.id AND approvals.approved_at IS NULL) > 0 THEN 'Pending Approval'
					WHEN (
						SELECT COUNT(*) FROM invoices 
						JOIN contracts ON contracts.id = invoices.contract_id AND contracts.company_id = companies.id
						WHERE invoices.status = 'Overdue'
					) > 5 THEN 'At Risk'
					ELSE 'Active'
				END
			`,
			Dependencies: map[string]domainDynamicColumn.Dependency{
				"companies": {
					Columns: []string{"is_active"},
				},
				"invoices": {
					Columns:           []string{"status", "contract_id"},
					RecordIdsSelector: "SELECT contract_id FROM invoices inv WHERE inv.id IN ({invoices.ids})",
				},
				"approvals": {
					Columns:           []string{"is_approved", "company_id"},
					RecordIdsSelector: "SELECT company_id FROM approvals ap WHERE ap.id IN ({approvals.ids})",
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
