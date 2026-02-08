-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS company (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_company_status ON company(status);
CREATE INDEX idx_company_is_deleted ON company(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS contract (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    company_id BIGINT NOT NULL,
    value DECIMAL(15,2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_contract_company FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE
);

CREATE INDEX idx_contract_company_id ON contract(company_id);
CREATE INDEX idx_contract_status ON contract(status);
CREATE INDEX idx_contract_is_deleted ON contract(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invoice (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(255) UNIQUE,
    description TEXT,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    pending_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    due_date TIMESTAMPTZ,
    status VARCHAR(50),
    payment_terms INTEGER NOT NULL DEFAULT 30,
    paid_at TIMESTAMPTZ,
    contract_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT chk_payment_terms_positive CHECK (payment_terms > 0),
    CONSTRAINT fk_invoice_contract FOREIGN KEY (contract_id) REFERENCES contract(id) ON DELETE CASCADE
);

CREATE INDEX idx_invoice_invoice_number ON invoice(invoice_number);
CREATE INDEX idx_invoice_status ON invoice(status);
CREATE INDEX idx_invoice_contract_id ON invoice(contract_id);
CREATE INDEX idx_invoice_is_deleted ON invoice(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payment (
    id BIGSERIAL PRIMARY KEY,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invoice_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_payment_invoice FOREIGN KEY (invoice_id) REFERENCES invoice(id) ON DELETE CASCADE
);

CREATE INDEX idx_payment_invoice_id ON payment(invoice_id);
CREATE INDEX idx_payment_is_deleted ON payment(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS approval (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    approver_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    comments TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_approval_company FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE,
    CONSTRAINT chk_approval_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_approval_company_id ON approval(company_id);
CREATE INDEX idx_approval_status ON approval(status);
CREATE INDEX idx_approval_is_deleted ON approval(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    position VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_employee_email ON employee(email);
CREATE INDEX idx_employee_status ON employee(status);
CREATE INDEX idx_employee_is_deleted ON employee(is_deleted);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS deployment (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    contract_id BIGINT NOT NULL,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    employee_id BIGINT UNIQUE,
    role VARCHAR(255),
    checkin_at TIMESTAMPTZ,
    checkout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_deployment_contract FOREIGN KEY (contract_id) REFERENCES contract(id) ON DELETE CASCADE,
    CONSTRAINT fk_deployment_employee FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE SET NULL
);

CREATE INDEX idx_deployment_contract_id ON deployment(contract_id);
CREATE INDEX idx_deployment_employee_id ON deployment(employee_id);
CREATE INDEX idx_deployment_status ON deployment(status);
CREATE INDEX idx_deployment_is_deleted ON deployment(is_deleted);
CREATE INDEX idx_deployment_is_cancelled ON deployment(is_cancelled);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dynamic_column (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    formula TEXT NOT NULL,
    default_value TEXT,
    type VARCHAR(50) NOT NULL,
    dependencies JSONB,
    CONSTRAINT unique_dynamic_column_table_name UNIQUE (table_name, name)
);

CREATE INDEX idx_dynamic_column_table_name ON dynamic_column(table_name);
CREATE INDEX idx_dynamic_column_name ON dynamic_column(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dynamic_column;
DROP TABLE IF EXISTS deployment;
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS approval;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS invoice;
DROP TABLE IF EXISTS contract;
DROP TABLE IF EXISTS company;
-- +goose StatementEnd
