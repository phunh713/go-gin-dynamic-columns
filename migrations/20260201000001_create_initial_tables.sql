-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(50),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_is_deleted ON companies(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS contracts (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    company_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_contracts_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

CREATE INDEX idx_contracts_company_id ON contracts(company_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_is_deleted ON contracts(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(255) UNIQUE,
    description TEXT,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    pending_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_date BIGINT,
    status VARCHAR(50),
    payment_terms INTEGER NOT NULL DEFAULT 30,
    paid_at BIGINT,
    contract_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT chk_payment_terms_positive CHECK (payment_terms > 0),
    CONSTRAINT fk_invoices_contract FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE
);

CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_contract_id ON invoices(contract_id);
CREATE INDEX idx_invoices_is_deleted ON invoices(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    paid_at BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_payments_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);

CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_is_deleted ON payments(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS approvals (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    approver_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    comments TEXT,
    reviewed_at BIGINT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_approvals_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT chk_approval_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_approvals_company_id ON approvals(company_id);
CREATE INDEX idx_approvals_status ON approvals(status);
CREATE INDEX idx_approvals_is_deleted ON approvals(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS deployments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    contract_id BIGINT NOT NULL,
    start_date BIGINT,
    end_date BIGINT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_deployments_contract FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE
);

CREATE INDEX idx_deployments_contract_id ON deployments(contract_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_is_deleted ON deployments(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    position VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_employees_email ON employees(email);
CREATE INDEX idx_employees_status ON employees(status);
CREATE INDEX idx_employees_is_deleted ON employees(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_deployments (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL,
    deployment_id BIGINT NOT NULL,
    role VARCHAR(255),
    created_at BIGINT NOT NULL,
    CONSTRAINT fk_employee_deployments_employee FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    CONSTRAINT fk_employee_deployments_deployment FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE,
    CONSTRAINT unique_employee_deployment UNIQUE (employee_id, deployment_id)
);

CREATE INDEX idx_employee_deployments_employee_id ON employee_deployments(employee_id);
CREATE INDEX idx_employee_deployments_deployment_id ON employee_deployments(deployment_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dynamic_columns (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    formula TEXT NOT NULL,
    default_value TEXT,
    type VARCHAR(50) NOT NULL,
    dependencies JSONB,
    CONSTRAINT unique_table_column UNIQUE (table_name, name)
);

CREATE INDEX idx_dynamic_columns_table_name ON dynamic_columns(table_name);
CREATE INDEX idx_dynamic_columns_name ON dynamic_columns(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dynamic_columns;
DROP TABLE IF EXISTS employee_deployments;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS approvals;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS companies;
-- +goose StatementEnd
