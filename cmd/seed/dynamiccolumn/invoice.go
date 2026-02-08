package dynamiccolumn

import (
	"fmt"
	"gin-demo/internal/domain/dynamiccolumn"
	domainDynamicColumn "gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"

	"gorm.io/gorm"
)

func seedInvoices(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "invoice",
			Name:      "pending_amount",
			Type:      "float",
			Formula: fmt.Sprintf(`
			WITH invoice_payments AS (
				SELECT inv.id,
					COALESCE(inv.total_amount - SUM(pmt.amount), inv.total_amount) AS pending_amount
				FROM invoice inv
				JOIN %s p ON inv.id = p.id
				LEFT JOIN payment pmt ON pmt.invoice_id = inv.id AND pmt.is_deleted = false
				WHERE inv.is_deleted = false
				GROUP BY inv.id
			)
			UPDATE invoice i
			SET pending_amount = ip.pending_amount
			FROM invoice_payments ip
			WHERE i.id = ip.id AND i.pending_amount IS DISTINCT FROM ip.pending_amount
			`, constants.TEMP_TABLE_NAME),
			Dependencies: map[constants.TableName]dynamiccolumn.Dependency{
				constants.TableNameInvoice: {
					Columns: []string{"total_amount"},
				},
				constants.TableNamePayment: {
					Columns:           []string{"amount", "invoice_id"},
					RecordIdsSelector: "SELECT invoice_id FROM payment WHERE payment.id IN ({payment.ids}) UNION SELECT {payment:original.invoice_id} as invoice_id",
				},
			},
		},
		{
			TableName: "invoice",
			Name:      "status",
			Type:      "string",
			Formula: fmt.Sprintf(`
				CASE 
					WHEN pending_amount <= 0 THEN '%s'
					WHEN CURRENT_DATE - created_at > payment_terms * INTERVAL '1 day' THEN '%s'
					ELSE '%s' 
				END
			`, constants.InvoiceStatusPaid, constants.InvoiceStatusOverdue, constants.InvoiceStatusPending),
			Dependencies: map[constants.TableName]domainDynamicColumn.Dependency{
				constants.TableNameInvoice: {
					Columns: []string{"pending_amount", "created_at", "payment_terms"},
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
