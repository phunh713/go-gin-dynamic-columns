package container

import (
	"gin-demo/internal/application/config"
	"gin-demo/internal/domain/approval"
	"gin-demo/internal/domain/company"
	"gin-demo/internal/domain/contract"
	"gin-demo/internal/domain/deployment"
	"gin-demo/internal/domain/dynamiccolumn"
	"gin-demo/internal/domain/employee"
	"gin-demo/internal/domain/invoice"
	"gin-demo/internal/domain/payment"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/types"
	"gin-demo/internal/shared/utils"
	"log/slog"
)

type Container struct {
	// Logger
	Logger *slog.Logger

	// Shared Dependencies can be added here
	DynamicColumnRepository dynamiccolumn.DynamicColumnRepository
	DynamicColumnService    dynamiccolumn.DynamicColumnService

	// Invoice Domain
	InvoiceRepository invoice.InvoiceRepository
	InvoiceService    invoice.InvoiceService
	InvoiceHandler    invoice.InvoiceHandler

	// Approval Domain
	ApprovalRepository approval.ApprovalRepository
	ApprovalService    approval.ApprovalService
	ApprovalHandler    approval.ApprovalHandler

	// Employee Domain
	EmployeeRepository employee.EmployeeRepository
	EmployeeService    employee.EmployeeService
	EmployeeHandler    employee.EmployeeHandler

	// Deployment Domain
	DeploymentRepository deployment.DeploymentRepository
	DeploymentService    deployment.DeploymentService
	DeploymentHandler    deployment.DeploymentHandler

	// Contract Domain
	ContractRepository contract.ContractRepository
	ContractService    contract.ContractService
	ContractHandler    contract.ContractHandler

	// Company Domain
	CompanyRepository company.CompanyRepository
	CompanyService    company.CompanyService
	CompanyHandler    company.CompanyHandler

	// Payment Domain
	PaymentRepository payment.PaymentRepository
	PaymentService    payment.PaymentService
	PaymentHandler    payment.PaymentHandler
}

func NewModelsMap() types.ModelsMap {
	return types.ModelsMap{
		//insert table models here
		constants.TableNameInvoice:    invoice.Invoice{},
		constants.TableNameContract:   contract.Contract{},
		constants.TableNameCompany:    company.Company{},
		constants.TableNamePayment:    payment.Payment{},
		constants.TableNameEmployee:   employee.Employee{},
		constants.TableNameApproval:   approval.Approval{},
		constants.TableNameDeployment: deployment.Deployment{},
	}
}

func NewContainer() *Container {
	c := &Container{}

	// Logger
	logger := config.NewLogger()
	c.Logger = logger

	// Shared Dependencies can be initialized here
	modelsMap := NewModelsMap()
	modelRelationsMap := utils.BuildRelationMap(modelsMap)

	c.DynamicColumnRepository = dynamiccolumn.NewDynamicColumnRepository(modelsMap, modelRelationsMap)
	c.DynamicColumnService = dynamiccolumn.NewDynamicColumnService(c.DynamicColumnRepository, modelsMap, modelRelationsMap, logger)

	// Invoice
	c.InvoiceRepository = invoice.NewInvoiceRepository()
	c.InvoiceService = invoice.NewInvoiceService(c.InvoiceRepository, c.DynamicColumnService)
	c.InvoiceHandler = invoice.NewInvoiceHandler(c.InvoiceService)

	// Approval
	c.ApprovalRepository = approval.NewApprovalRepository()
	c.ApprovalService = approval.NewApprovalService(c.ApprovalRepository, c.DynamicColumnService)
	c.ApprovalHandler = approval.NewApprovalHandler(c.ApprovalService)

	// Employee
	c.EmployeeRepository = employee.NewEmployeeRepository()
	c.EmployeeService = employee.NewEmployeeService(c.EmployeeRepository, c.DynamicColumnService)
	c.EmployeeHandler = employee.NewEmployeeHandler(c.EmployeeService)

	// Deployment
	c.DeploymentRepository = deployment.NewDeploymentRepository()
	c.DeploymentService = deployment.NewDeploymentService(c.DeploymentRepository, c.DynamicColumnService)
	c.DeploymentHandler = deployment.NewDeploymentHandler(c.DeploymentService)

	// Contract
	c.ContractRepository = contract.NewContractRepository()
	c.ContractService = contract.NewContractService(c.ContractRepository, c.DynamicColumnService)
	c.ContractHandler = contract.NewContractHandler(c.ContractService)

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
