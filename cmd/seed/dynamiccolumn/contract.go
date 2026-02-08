package dynamiccolumn

import (
	"fmt"
	domainDynamicColumn "gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"

	"gorm.io/gorm"
)

func seedContracts(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "contract",
			Name:      "status",
			Type:      "string",
			Formula: fmt.Sprintf(`
				WITH contract_invoices AS (
					SELECT inv.contract_id,
						COUNT(*) FILTER (WHERE inv.status = '%s') as overdue_count,
						COUNT(*) as total_count
					FROM invoice inv
					JOIN %s p ON inv.contract_id = p.id
					WHERE inv.is_deleted = false 
					GROUP BY inv.contract_id
				),
				contract_deployments AS (
					SELECT dep.contract_id,
						COUNT(*) as total_count,
						COUNT(*) FILTER (WHERE dep.status <> '%s') as non_completed_count
					FROM deployment dep
					JOIN %s p ON dep.contract_id = p.id
					WHERE dep.is_deleted = false
					GROUP BY dep.contract_id
				),
				contract_companies AS (
					SELECT ct.id as contract_id,
						cp.status
					FROM contract ct
					JOIN %s p ON ct.id = p.id
					JOIN company cp ON cp.id = ct.company_id AND cp.is_deleted = false
				),
				contract_status AS (
					SELECT ct.id,
						CASE
							WHEN ct.is_cancelled = true THEN '%s'
							WHEN cp.status <> '%s' THEN '%s'
							WHEN CURRENT_DATE < ct.start_date THEN '%s'
							WHEN CURRENT_DATE > ct.end_date THEN
								CASE
									WHEN COALESCE(cd.total_count, 0) = 0 THEN '%s'
									WHEN cd.non_completed_count > 0 THEN '%s'
									WHEN COALESCE(ci.total_count, 0) = 0 THEN '%s'
									WHEN ci.overdue_count > 0 THEN '%s'
									ELSE '%s'
								END
							WHEN COALESCE(cd.total_count, 0) = 0 THEN '%s'
							ELSE '%s'
						END AS status
					FROM contract ct
					JOIN %s p ON ct.id = p.id
					LEFT JOIN contract_invoices ci ON ci.contract_id = ct.id
					LEFT JOIN contract_deployments cd ON cd.contract_id = ct.id
					LEFT JOIN contract_companies cp ON cp.contract_id = ct.id
				)
				UPDATE contract c
				SET status = cs.status
				FROM contract_status cs
				WHERE c.id = cs.id AND c.status IS DISTINCT FROM cs.status
			`,
				constants.InvoiceStatusOverdue, constants.TEMP_TABLE_NAME, constants.DeploymentStatusCompleted,
				constants.TEMP_TABLE_NAME, constants.TEMP_TABLE_NAME, constants.ContractStatusCanceled,
				constants.CompanyStatusActive, constants.ContractStatusCompanyNotActive, constants.ContractStatusInitiated,
				constants.ContractStatusExpiredNoDeployment, constants.ContractStatusDeploymentPending, constants.ContractStatusNoInvoice,
				constants.ContractStatusInvoiceOverdue, constants.ContractStatusCompleted, constants.ContractStatusNoDeployment,
				constants.ContractStatusActive, constants.TEMP_TABLE_NAME,
			),
			Dependencies: map[constants.TableName]domainDynamicColumn.Dependency{
				constants.TableNameContract: {
					Columns: []string{"is_cancelled", "start_date", "end_date"},
				},
				constants.TableNameCompany: {
					Columns:           []string{"status"},
					RecordIdsSelector: "SELECT id FROM contract WHERE company_id IN ({company.ids})",
				},
				constants.TableNameInvoice: {
					Columns:           []string{"status", "contract_id", "is_deleted"},
					RecordIdsSelector: "SELECT contract_id FROM invoice WHERE invoice.id IN ({invoice.ids})",
				},
				constants.TableNameDeployment: {
					Columns:           []string{"contract_id", "is_deleted", "status"},
					RecordIdsSelector: "SELECT contract_id FROM deployment WHERE deployment.id IN ({deployment.ids})",
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
