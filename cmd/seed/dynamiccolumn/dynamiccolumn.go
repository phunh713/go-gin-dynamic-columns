package dynamiccolumn

import (
	"log/slog"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, logger *slog.Logger) {
	seedCompanies(db, logger)
	seedInvoices(db, logger)
	seedContracts(db, logger)
	seedDeployments(db, logger)
}
