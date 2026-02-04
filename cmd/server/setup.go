package main

import (
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/company"
	"gin-demo/internal/domain/invoice"
	"gin-demo/internal/domain/payment"
	"gin-demo/internal/shared/base"
)

func SetupRoutes(app *config.App, c *container.Container) {
	invoice.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.InvoiceHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.InvoiceHandler.GetById},
		{Method: "POST", Path: "", Handler: c.InvoiceHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.InvoiceHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.InvoiceHandler.Delete},
	})
	company.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.CompanyHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.CompanyHandler.GetById},
		{Method: "POST", Path: "", Handler: c.CompanyHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.CompanyHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.CompanyHandler.Delete},
	})
	payment.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.PaymentHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.PaymentHandler.GetById},
		{Method: "POST", Path: "", Handler: c.PaymentHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.PaymentHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.PaymentHandler.Delete},
	})
}
