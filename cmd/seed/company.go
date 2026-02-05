package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/company"
	"time"

	"gorm.io/gorm"
)

func SeedCompanies(db *gorm.DB) {
	totalStart := time.Now()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	container := container.NewContainer()

	for i := 1; i <= 50; i++ {
		start := time.Now()
		company := company.Company{
			Name:     "Company A " + fmt.Sprint(i),
			IsActive: true,
		}
		container.CompanyService.Create(ctx, &company)
		elapsed := time.Since(start).Seconds()
		fmt.Printf("Created company ID %d in %.4f seconds\n", company.Id, elapsed)
	}

	totalElapsed := time.Since(totalStart).Seconds()
	fmt.Printf("\n✓ Total time for SeedCompanies: %.4f seconds\n", totalElapsed)
}
