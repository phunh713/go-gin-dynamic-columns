-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    is_working BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_created_at ON companies(created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    pending_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL,
    payment_terms INTEGER NOT NULL DEFAULT 30,
    company_id BIGINT NOT NULL,
    force_payment BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT chk_payment_terms_positive CHECK (payment_terms > 0),
    CONSTRAINT chk_total_amount_positive CHECK (total_amount >= 0)
);

CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_paid_at ON invoices(paid_at);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);
CREATE INDEX idx_invoices_payment_terms ON invoices(payment_terms);
CREATE INDEX idx_invoices_company_id ON invoices(company_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invoice_id BIGINT NOT NULL,
    CONSTRAINT chk_payment_amount_positive CHECK (amount > 0),
    CONSTRAINT fk_payments_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);

CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_paid_at ON payments(paid_at);
CREATE INDEX idx_payments_created_at ON payments(created_at);
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
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS companies;
-- +goose StatementEnd
