-- Migration 000001: Socle transactionnel persistant (PostgreSQL natif)

CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country_code VARCHAR(2) NOT NULL DEFAULT 'FR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_credentials (
    key_id VARCHAR(32) PRIMARY KEY,
    key_prefix VARCHAR(16) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    scopes TEXT[] NOT NULL DEFAULT '{"invoices:write", "invoices:read"}',
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_api_credentials_tenant ON api_credentials(tenant_id);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id VARCHAR(64) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(128) NOT NULL,
    request_hash CHAR(64) NOT NULL,
    invoice_id VARCHAR(64),
    response_status INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, key)
);
CREATE INDEX IF NOT EXISTS idx_idempotency_expiry ON idempotency_keys(expires_at);

CREATE TABLE IF NOT EXISTS invoices (
    tenant_id VARCHAR(64) NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    id VARCHAR(64) NOT NULL,
    invoice_number VARCHAR(128) NOT NULL,
    seller_identifier VARCHAR(64) NOT NULL,
    buyer_identifier VARCHAR(64) NOT NULL,
    issue_date DATE NOT NULL,
    currency VARCHAR(3) NOT NULL,
    total_tax_inclusive NUMERIC(15, 2) NOT NULL,
    syntax VARCHAR(32) NOT NULL,
    profile VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    document_sha256 CHAR(64) NOT NULL,
    document_storage_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT uq_tenant_invoice_business UNIQUE (tenant_id, seller_identifier, invoice_number, issue_date),
    CONSTRAINT uq_tenant_document_sha256 UNIQUE (tenant_id, document_sha256)
);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(tenant_id, status);

CREATE TABLE IF NOT EXISTS invoice_events (
    tenant_id VARCHAR(64) NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    invoice_id VARCHAR(64) NOT NULL,
    sequence INT NOT NULL,
    event_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor VARCHAR(64) NOT NULL,
    document_sha256 CHAR(64) NOT NULL,
    payload_summary TEXT,
    previous_hash CHAR(64) NOT NULL,
    current_hash CHAR(64) NOT NULL,
    timestamp_utc TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, invoice_id, sequence)
);
