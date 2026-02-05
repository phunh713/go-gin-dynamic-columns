package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/approval"
	"time"

	"gorm.io/gorm"
)

func SeedApprovals(db *gorm.DB) {
	totalStart := time.Now()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	container := container.NewContainer()

	for i := 11; i <= 20; i++ {
		start := time.Now()
		approval := approval.Approval{
			CompanyId:    int64(i),
			ApproverName: fmt.Sprintf("Approver %d", i),
			ReviewedAt:   &time.Time{},
			Comments:     "Passed",
		}
		container.ApprovalService.Create(ctx, &approval)
		elapsed := time.Since(start).Seconds()
		fmt.Printf("Created approval ID %d in %.4f seconds\n", approval.Id, elapsed)
	}

	totalElapsed := time.Since(totalStart).Seconds()
	fmt.Printf("\n✓ Total time for SeedApproval: %.4f seconds\n", totalElapsed)
}
