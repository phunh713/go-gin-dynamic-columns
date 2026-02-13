package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/contract"
	"time"

	"gorm.io/gorm"
)

func SeedContracts(db *gorm.DB) {
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	container := container.NewContainer()
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	companies := container.CompanyService.GetAllCompanies(ctx)
	for _, comp := range companies {
		fmt.Printf("Seeding contracts for company ID %d...\n", comp.Id)
		totalStart := time.Now()
		contracts := make([]contract.Contract, 0)
		for i := 1; i <= 5000; i++ {
			contract := contract.Contract{
				Name:        "Contract C " + fmt.Sprint(i),
				CompanyId:   1,
				Description: "Contract No." + fmt.Sprint(i),
				IsCancelled: false,
				Value:       50000 + float64(i)*float64(20.5),
				StartDate:   time.Now().AddDate(0, 0, i-500),
				EndDate:     time.Now().AddDate(0, 6, i-300),
			}
			contracts = append(contracts, contract)
		}

		// Seed 300 cancelled contracts per company
		for i := 1; i <= 1500; i++ {
			contract := contract.Contract{
				Name:        "Contract C " + fmt.Sprint(i),
				CompanyId:   1,
				Description: "Contract No." + fmt.Sprint(i),
				IsCancelled: true,
				Value:       50000 + float64(i)*float64(20.5),
				StartDate:   time.Now().AddDate(0, 0, i-500),
				EndDate:     time.Now().AddDate(0, 6, i-300),
			}
			contracts = append(contracts, contract)
		}

		tx := db.Begin()
		ctx = context.WithValue(ctx, config.ContextKeyDB, tx)
		container.ContractService.CreateMultiple(ctx, contracts)
		tx.Commit()
		totalElapsed := time.Since(totalStart).Seconds()
		fmt.Printf("\n✓ Total time for SeedContracts: %.4f seconds\n", totalElapsed)
	}
}
