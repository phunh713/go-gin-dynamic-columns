package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/approval"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

func SeedApprovals(db *gorm.DB, logger *slog.Logger) {
	totalStart := time.Now()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	container := container.NewContainer()

	for i := 1; i <= 30; i++ {
		tx := db.Begin()
		ctx = context.WithValue(ctx, config.ContextKeyDB, tx)
		start := time.Now()
		approval := approval.Approval{
			CompanyId:    int64(i),
			ApproverName: fmt.Sprintf("Approver %d", i),
			ReviewedAt:   &time.Time{},
			Comments:     "Passed",
			Status:       "approved",
		}
		container.ApprovalService.Create(ctx, &approval)
		elapsed := time.Since(start).Seconds()
		fmt.Printf("Created approval ID %d in %.4f seconds\n", approval.Id, elapsed)
		tx.Commit()
	}

	totalElapsed := time.Since(totalStart).Seconds()
	fmt.Printf("\n✓ Total time for SeedApproval: %.4f seconds\n", totalElapsed)
}
