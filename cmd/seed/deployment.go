package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/deployment"
	"time"

	"gorm.io/gorm"
)

func SeedDeployments(db *gorm.DB) {
	totalStart := time.Now()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	container := container.NewContainer()
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	contracts := container.ContractService.GetAll(ctx)
	fmt.Println(len(contracts))
	for _, contract := range contracts {
		for i := 1; i <= 10; i++ {
			tx := db.Begin()
			ctx = context.WithValue(ctx, config.ContextKeyDB, tx)
			start := time.Now()
			deployment := deployment.Deployment{
				Name:        "Deployment A " + fmt.Sprint(i),
				ContractId:  contract.Id,
				Description: "Something Fun",
				IsCancelled: false,
			}
			container.DeploymentService.Create(ctx, &deployment)
			tx.Commit()
			elapsed := time.Since(start).Seconds()
			fmt.Printf("Created deployment ID %d in %.4f seconds\n", deployment.Id, elapsed)
		}
	}
	totalElapsed := time.Since(totalStart).Seconds()
	fmt.Printf("\n✓ Total time for SeedDeployments: %.4f seconds\n", totalElapsed)
}
