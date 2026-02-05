package main

import (
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/approval"
	"gin-demo/internal/domain/company"
	"gin-demo/internal/domain/contract"
	"gin-demo/internal/domain/deployment"
	"gin-demo/internal/domain/employee"
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
	approval.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.ApprovalHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.ApprovalHandler.GetById},
		{Method: "POST", Path: "", Handler: c.ApprovalHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.ApprovalHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.ApprovalHandler.Delete},
	})
	contract.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.ContractHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.ContractHandler.GetById},
		{Method: "POST", Path: "", Handler: c.ContractHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.ContractHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.ContractHandler.Delete},
	})
	employee.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.EmployeeHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.EmployeeHandler.GetById},
		{Method: "POST", Path: "", Handler: c.EmployeeHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.EmployeeHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.EmployeeHandler.Delete},
	})
	deployment.RegisterRoutes("v1", app, []base.HandlerConfig{
		{Method: "GET", Path: "", Handler: c.DeploymentHandler.GetAll},
		{Method: "GET", Path: "/:id", Handler: c.DeploymentHandler.GetById},
		{Method: "POST", Path: "", Handler: c.DeploymentHandler.Create},
		{Method: "PUT", Path: "/:id", Handler: c.DeploymentHandler.Update},
		{Method: "DELETE", Path: "/:id", Handler: c.DeploymentHandler.Delete},
	})
}
