package container

import (
	"gin-demo/internal/domain/company"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/domain/invoice"
	"gin-demo/internal/domain/payment"
)

type Container struct {
	// Shared Dependencies can be added here
	DynamicColumnRepository dynamiccolumn.DynamicColumnRepository
	DynamicColumnService    dynamiccolumn.DynamicColumnService

	// Invoice Domain
	InvoiceRepository invoice.InvoiceRepository
	InvoiceService    invoice.InvoiceService
	InvoiceHandler    invoice.InvoiceHandler

	// Company Domain
	CompanyRepository company.CompanyRepository
	CompanyService    company.CompanyService
	CompanyHandler    company.CompanyHandler

	// Payment Domain
	PaymentRepository payment.PaymentRepository
	PaymentService    payment.PaymentService
	PaymentHandler    payment.PaymentHandler
}

func NewModelsMap() map[string]interface{} {
	return map[string]interface{}{
		//insert table models here
		"invoices":  invoice.Invoice{},
		"companies": company.Company{},
		"payments":  payment.Payment{},
	}
}

func NewContainer() *Container {
	c := &Container{}

	// Shared Dependencies can be initialized here
	modelsMap := NewModelsMap()
	c.DynamicColumnRepository = dynamiccolumn.NewDynamicColumnRepository(modelsMap)
	c.DynamicColumnService = dynamiccolumn.NewDynamicColumnService(c.DynamicColumnRepository, modelsMap)

	// Invoice
	c.InvoiceRepository = invoice.NewInvoiceRepository()
	c.InvoiceService = invoice.NewInvoiceService(c.InvoiceRepository, c.DynamicColumnService)
	c.InvoiceHandler = invoice.NewInvoiceHandler(c.InvoiceService)

	// Company
	c.CompanyRepository = company.NewCompanyRepository()
	c.CompanyService = company.NewCompanyService(c.CompanyRepository, c.DynamicColumnService)
	c.CompanyHandler = company.NewCompanyHandler(c.CompanyService)

	// Payment
	c.PaymentRepository = payment.NewPaymentRepository()
	c.PaymentService = payment.NewPaymentService(c.PaymentRepository, c.DynamicColumnService)
	c.PaymentHandler = payment.NewPaymentHandler(c.PaymentService)

	return c
}
