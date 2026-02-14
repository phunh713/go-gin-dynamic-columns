package dynamiccolumn

import (
	"fmt"
	"gin-demo/internal/shared/constants"
	domainDynamicColumn "gin-demo/internal/system/dynamiccolumn"
	"log/slog"

	"gorm.io/gorm"
)

func seedCompanies(db *gorm.DB, logger *slog.Logger) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "company",
			Name:      "status",
			Type:      "string",
			Formula: fmt.Sprintf(`
				WITH company_invoices AS (
					SELECT 
						ct.company_id,
						COUNT(*) FILTER (WHERE inv.status = '%s') as overdue_count
					FROM invoice inv
					JOIN contract ct ON ct.id = inv.contract_id AND ct.is_deleted = false
					JOIN %s p ON ct.company_id = p.id
					WHERE inv.is_deleted = false 
					GROUP BY ct.company_id
				),
				company_approvals AS (
					SELECT 
						company_id,
						COUNT(*) as total_count,
						COUNT(*) FILTER (WHERE status <> '%s') as non_approved_count
					FROM approval
					JOIN %s p ON company_id = p.id
					WHERE is_deleted = false 
					GROUP BY company_id
				),
				company_status AS (
					SELECT 
						c.id,
						CASE
							WHEN c.is_active = false THEN '%s'
							WHEN COALESCE(ca.total_count, 0) = 0 THEN '%s'
							WHEN COALESCE(ca.non_approved_count, 0) > 0 THEN '%s'
							WHEN COALESCE(ci.overdue_count, 0) > 5 THEN '%s'
							ELSE '%s'
						END as status
					FROM company c
					JOIN %s p ON c.id = p.id
					LEFT JOIN company_invoices ci ON ci.company_id = c.id
					LEFT JOIN company_approvals ca ON ca.company_id = c.id
				)
				UPDATE company cp
				SET status = cs.status
				FROM company_status cs
				WHERE cp.id = cs.id
				AND cp.status IS DISTINCT FROM cs.status;
			`,
				constants.InvoiceStatusOverdue, constants.TEMP_TABLE_NAME, constants.ApprovalApproved,
				constants.TEMP_TABLE_NAME, constants.CompanyStatusInactive, constants.CompanyStatusNoApproval,
				constants.CompanyStatusPending, constants.CompanyStatusAtRisk, constants.CompanyStatusActive,
				constants.TEMP_TABLE_NAME),
			Dependencies: map[constants.TableName]domainDynamicColumn.Dependency{
				constants.TableNameCompany: {
					Columns: []string{"is_active"},
				},
				constants.TableNameContract: {
					Columns:           []string{"company_id", "is_deleted"},
					RecordIdsSelector: "SELECT company_id FROM contract WHERE contract.id IN ({contract.ids})",
				},
				constants.TableNameInvoice: {
					Columns:           []string{"status", "contract_id", "is_deleted"},
					RecordIdsSelector: "SELECT company_id FROM contract JOIN invoice ON contract.id = invoice.contract_id AND invoice.id IN ({invoice.ids})",
				},
				constants.TableNameApproval: {
					Columns:           []string{"status", "company_id", "is_deleted"},
					RecordIdsSelector: "SELECT company_id FROM approval ap WHERE ap.id IN ({approval.ids})",
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

const VARS = `
{{invoice}}.overdue_count = COUNT(*) FILTER (WHERE {{invoice}}.status = 'Overdue')
{{contract}}.overdue_count = COUNT
`
