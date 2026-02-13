package main

import (
	"context"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/domain/dynamiccolumn"
)

func main() {
	// Load config
	configEnv := config.LoadEnv()

	// Connect to database
	db := config.NewDB(configEnv)
	c := container.NewContainer()
	ctx := context.Background()
	// add db to ctx so that it can be used in service/repository layers
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)

	contract_status := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "contract",
		Name:      "status",
		Formula: `
			CASE
				WHEN {{contract}}.is_cancelled = true THEN 'Canceled'
				WHEN {{company}}.status <> 'Active' THEN 'On Hold - Company Not Active'
				WHEN CURRENT_DATE < {{contract}}.start_date THEN 'Initiated'
				WHEN CURRENT_DATE > {{contract}}.end_date THEN
					CASE
						WHEN COALESCE(deployment_total_count, 0) = 0 THEN 'Expired - No Deployment'
						WHEN deployment_non_completed_count > 0 THEN 'Expired - Deployment Pending'
						WHEN COALESCE(invoice_total_count, 0) = 0 THEN 'Completed - No Invoice'
						WHEN invoice_overdue_count > 0 THEN 'Expired - Invoice Overdue'
						ELSE 'Completed'
					END
				WHEN COALESCE(deployment_total_count, 0) = 0 THEN 'Need Attention - No Deployment'
				ELSE 'Active'
			END
		`,
		Variables: `
			var deployment_non_completed_count = COUNT({{deployment}}.id) FILTER (WHERE {{deployment}}.status <> 'Completed')
			var deployment_total_count = COUNT({{deployment}}.id)
			var invoice_total_count = COUNT({{invoice}}.id)
			var invoice_overdue_count = COUNT({{invoice}}.id) FILTER (WHERE {{invoice}}.status = 'Overdue')
		`,
		Type: "string",
	}

	company_status := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "company",
		Name:      "status",
		Formula: `
			CASE
				WHEN {{company}}.is_active = false THEN 'Inactive'
				WHEN COALESCE(approval_total_count, 0) = 0 THEN 'No Approval'
				WHEN COALESCE(approval_non_approved_count, 0) > 0 THEN 'Pending Approval'
				WHEN COALESCE(invoice_overdue_count, 0) > 5 THEN 'At Risk'
				ELSE 'Active'
			END
		`,
		Variables: `
			var invoice_overdue_count = COUNT({{invoice}}.id) FILTER (WHERE {{invoice}}.status = 'Overdue')
			var approval_total_count = COUNT({{approval}}.id)
			var approval_non_approved_count = COUNT({{approval}}.id) FILTER (WHERE {{approval}}.status <> 'approved')
		`,
		Type: "string",
	}

	invoice_pending_amount := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "invoice",
		Name:      "pending_amount",
		Type:      "float",
		Formula: `
			COALESCE({{invoice}}.total_amount - payment_total_amount, {{invoice}}.total_amount)
		`,
		Variables: `
			var payment_total_amount = SUM({{payment}}.amount)
		`,
	}

	invoice_status := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "invoice",
		Name:      "status",
		Type:      "string",
		Formula: `
			CASE 
				WHEN {{invoice}}.pending_amount <= 0 THEN 'Paid'
				WHEN CURRENT_DATE - {{invoice}}.created_at > {{invoice}}.payment_terms * INTERVAL '1 day' THEN 'Overdue'
				ELSE 'Pending' 
			END
		`,
		Variables: ``,
	}

	deployment_status := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "deployment",
		Name:      "status",
		Type:      "string",
		Formula: `
			CASE
				WHEN {{deployment}}.is_cancelled = true THEN 'Canceled'
				WHEN {{deployment}}.employee_id IS NULL OR ({{deployment}}.checkin_at IS NULL AND {{deployment}}.checkout_at IS NULL) THEN 'Pending'
				WHEN {{deployment}}.checkout_at IS NULL THEN 'In Progress'
				ELSE 'Completed'
			END
		`,
		Variables: ``,
	}

	deployment_can_start := &dynamiccolumn.DynamicColumnCreateRequest{
		TableName: "deployment",
		Name:      "can_start",
		Type:      "bool",
		Formula: `
			CASE
				WHEN {{contract}}.status <> 'Active' THEN false
				WHEN {{deployment}}.is_cancelled = true THEN false
				ELSE true
			END
		`,
	}

	payload := []*dynamiccolumn.DynamicColumnCreateRequest{
		contract_status,
		company_status,
		invoice_pending_amount,
		invoice_status,
		deployment_status,
		deployment_can_start,
	}

	for _, p := range payload {
		c.DynamicColumnService.Create(ctx, p)
	}

	// Pretty print the formula please with JSON indent
	// formattedFormula, _ := json.MarshalIndent(formula, "", "  ")
	// fmt.Println(string(formattedFormula))

}
