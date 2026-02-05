package dynamiccolumn

import "gorm.io/gorm"

func Seed(db *gorm.DB) {
	seedCompanies(db)
	seedInvoices(db)
}
