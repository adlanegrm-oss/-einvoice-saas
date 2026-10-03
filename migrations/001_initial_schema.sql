-- 001_initial_schema.sql : Schema strict PostgreSQL & SQLite compatible
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    invoice_number VARCHAR(128) NOT NULL,
    compliance_status VARCHAR(32) NOT NULL,
    transmission_status VARCHAR(32) NOT NULL,
    payment_status VARCHAR(32) NOT NULL,
    total_cents BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT uq_tenant_invoice UNIQUE (tenant_id, invoice_number)
);

CREATE TABLE IF NOT EXISTS audit_events (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    invoice_id VARCHAR(64) NOT NULL,
    sequence_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    prev_hash VARCHAR(64) NOT NULL,
    event_hash VARCHAR(64) NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    CONSTRAINT uq_audit_seq UNIQUE (invoice_id, sequence_id)
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload_json TEXT NOT NULL,
    status VARCHAR(32) NOT NULL, -- PENDING, PROCESSING, SUCCESS, FAILED, DLQ
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    next_attempt_at TIMESTAMP NOT NULL,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    response_code INT,
    response_body TEXT,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, idempotency_key)
);
