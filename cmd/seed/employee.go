package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/employee"
	"time"

	"gorm.io/gorm"
)

func SeedEmployees(db *gorm.DB) {
	totalStart := time.Now()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	container := container.NewContainer()

	for i := 1; i <= 50; i++ {
		tx := db.Begin()
		ctx = context.WithValue(ctx, config.ContextKeyDB, tx)
		start := time.Now()
		company := employee.Employee{
			Name:  "Employee AE " + fmt.Sprint(i),
			Email: fmt.Sprintf("eployee%d@gmail.com", i),
		}
		container.EmployeeService.Create(ctx, &company)
		tx.Commit()
		elapsed := time.Since(start).Seconds()
		fmt.Printf("Created company ID %d in %.4f seconds\n", company.Id, elapsed)
	}
	totalElapsed := time.Since(totalStart).Seconds()
	fmt.Printf("\n✓ Total time for SeedCompanies: %.4f seconds\n", totalElapsed)
}
