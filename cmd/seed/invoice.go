package main

import (
	"context"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/invoice"
	"gin-demo/internal/shared/types"
	"time"

	"gorm.io/gorm"
)

func SeedInvoices(db *gorm.DB) {
	startTotal := time.Now()

	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	container := container.NewContainer()

	// container.DynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, "invoices", ids, constants.ActionRefresh, nil, nil, nil)
	invoices := make([]invoice.Invoice, 0)
	// for i := 1; i <= 50; i++ {
	for j := 1; j <= 6; j++ {
		createdAt := time.Date(2026, 02, 05, 0, 0, 0, 0, time.UTC)
		invoice := invoice.Invoice{
			ContractId:   int64(14),
			PaymentTerms: j * 5,
			TotalAmount:  float64(100 * j),
			GormModel: types.GormModel{
				CreatedAt: createdAt,
			},
		}
		invoices = append(invoices, invoice)
	}
	// }
	container.InvoiceService.CreateMultiple(ctx, invoices)

	totalElapsed := time.Since(startTotal).Seconds()
	fmt.Printf("\n✓ Total time for SeedInvoices: %.4f seconds\n", totalElapsed)
}
