package constants

type Action string

const (
	ActionCreate  Action = "CREATE"
	ActionUpdate  Action = "UPDATE"
	ActionDelete  Action = "DELETE"
	ActionRefresh Action = "REFRESH"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type CompanyStatus string

const (
	CompanyStatusActive     CompanyStatus = "Active"
	CompanyStatusAtRisk     CompanyStatus = "At Risk"
	CompanyStatusPending    CompanyStatus = "Pending Approval"
	CompanyStatusNoApproval CompanyStatus = "No Approval"
	CompanyStatusInactive   CompanyStatus = "Inactive"
)

type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "Pending"
	InvoiceStatusPaid    InvoiceStatus = "Paid"
	InvoiceStatusOverdue InvoiceStatus = "Overdue"
)

type ContractStatus string

const (
	ContractStatusCompanyNotActive    ContractStatus = "On Hold - Company Not Active"   // company is not active
	ContractStatusInitiated           ContractStatus = "Initiated"                      // before start date
	ContractStatusNoDeployment        ContractStatus = "Need Attention - No Deployment" // started but no deployment
	ContractStatusActive              ContractStatus = "Active"                         // started and has deployment
	ContractStatusExpiredNoDeployment ContractStatus = "Expired - No Deployment"        // past end date and no deployment
	ContractStatusInvoiceOverdue      ContractStatus = "Expired - Invoice Overdue"      // past end date and debt not cleared
	ContractStatusDeploymentPending   ContractStatus = "Expired - Deployment Pending"   // past end date and deployment not checked out/pending
	ContractStatusNoInvoice           ContractStatus = "Completed - No Invoice"         // past end date and debt cleared
	ContractStatusCompleted           ContractStatus = "Completed"                      // past end date and debt cleared
	ContractStatusCanceled            ContractStatus = "Canceled"
)

type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "Pending"     // No employee assigned
	DeploymentStatusInProgress DeploymentStatus = "In Progress" // Employee checked in
	DeploymentStatusCompleted  DeploymentStatus = "Completed"   // Employee checked out
	DeploymentStatusCanceled   DeploymentStatus = "Canceled"    // deployment.is_cancelled = true
)

type TableRelation string

const (
	TableRelationOneToMany  TableRelation = "one_to_many"
	TableRelationManyToOne  TableRelation = "many_to_one"
	TableRelationManyToMany TableRelation = "many_to_many"
	TableRelationNotRelated TableRelation = "not_related"
)
