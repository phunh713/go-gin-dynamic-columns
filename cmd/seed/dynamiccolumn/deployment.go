package dynamiccolumn

import (
	"fmt"
	"gin-demo/internal/domain/dynamiccolumn"
	domainDynamicColumn "gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/shared/constants"

	"gorm.io/gorm"
)

func seedDeployments(db *gorm.DB) {
	// Implement invoice seeding logic here if needed
	dycol := []domainDynamicColumn.DynamicColumn{
		{
			TableName: "deployment",
			Name:      "status",
			Type:      "string",
			Formula: fmt.Sprintf(`
			WITH deployment_status AS (
				SELECT dpl.id,
					CASE
						WHEN dpl.is_cancelled = true THEN '%s'
						WHEN dpl.employee_id IS NULL OR (dpl.checkin_at IS NULL AND dpl.checkout_at IS NULL) THEN '%s'
						WHEN dpl.checkout_at IS NULL THEN '%s'
						ELSE '%s'
					END as status
				FROM deployment dpl
				JOIN %s p ON dpl.id = p.id
				WHERE dpl.is_deleted = false
			)
			UPDATE deployment d
			SET status = ds.status
			FROM deployment_status ds
			WHERE ds.id = d.id AND d.status IS DISTINCT FROM ds.status
			`,
				constants.DeploymentStatusCanceled, constants.DeploymentStatusPending, constants.DeploymentStatusInProgress,
				constants.DeploymentStatusCompleted, constants.TEMP_TABLE_NAME),
			Dependencies: map[constants.TableName]dynamiccolumn.Dependency{
				constants.TableNameDeployment: {
					Columns: []string{"is_cancelled", "employee_id", "checkin_at", "checkout_at"},
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
