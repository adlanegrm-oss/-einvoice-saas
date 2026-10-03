-- Schema d'infrastructure E-Invoicing P0
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Organisations & Tenants
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(64) NOT NULL UNIQUE,
    legal_name VARCHAR(255) NOT NULL,
    siren_siret VARCHAR(14) NOT NULL,
    country_code VARCHAR(2) NOT NULL DEFAULT 'FR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Clés API d'intégration B2B
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key_prefix VARCHAR(16) NOT NULL,
    hashed_secret VARCHAR(128) NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL DEFAULT '{"invoice:read", "invoice:write"}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ
);
CREATE INDEX idx_api_keys_lookup ON api_keys(key_prefix, revoked_at);

-- 3. Idempotence Persistante
CREATE TABLE idempotency_keys (
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(128) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    response_code INT,
    response_body JSONB,
    invoice_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, idempotency_key)
);
CREATE INDEX idx_idempotency_expiry ON idempotency_keys(expires_at);

-- 4. Registre Factures Multi-Tenant
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invoice_number VARCHAR(128) NOT NULL,
    direction VARCHAR(16) NOT NULL DEFAULT 'OUTBOUND',
    profile VARCHAR(64) NOT NULL,
    compliance_status VARCHAR(32) NOT NULL DEFAULT 'NOT_VALIDATED',
    transmission_status VARCHAR(32) NOT NULL DEFAULT 'NOT_SUBMITTED',
    payment_status VARCHAR(32) NOT NULL DEFAULT 'NOT_DUE',
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    net_amount_cents BIGINT NOT NULL,
    tax_amount_cents BIGINT NOT NULL,
    gross_amount_cents BIGINT NOT NULL,
    buyer_identifier VARCHAR(64) NOT NULL,
    seller_identifier VARCHAR(64) NOT NULL,
    raw_payload_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_invoice_number UNIQUE (tenant_id, invoice_number)
);
CREATE INDEX idx_invoices_tenant_created ON invoices(tenant_id, created_at DESC);
CREATE INDEX idx_invoices_tenant_status ON invoices(tenant_id, transmission_status, compliance_status);

-- 5. Journal d'Audit Immuable
CREATE TABLE invoice_audit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    sequence_id BIGSERIAL NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload_hash CHAR(64) NOT NULL,
    prev_event_hash CHAR(64) NOT NULL,
    event_hash CHAR(64) NOT NULL,
    metadata JSONB,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_invoice_seq UNIQUE (invoice_id, sequence_id)
);
CREATE INDEX idx_audit_chain ON invoice_audit_events(invoice_id, sequence_id ASC);

-- 6. Transactional Outbox
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_outbox_queue ON outbox_events(status, next_attempt_at) 
WHERE status IN ('PENDING', 'FAILED');

-- 7. Delivery Ledger
CREATE TABLE delivery_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    transport_type VARCHAR(32) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    correlation_id VARCHAR(255),
    recipient_routing_id VARCHAR(128) NOT NULL,
    attempt_count INT NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL,
    nrr_evidence_hash VARCHAR(64),
    raw_response_payload TEXT,
    first_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);
CREATE INDEX idx_delivery_lookup ON delivery_ledger(tenant_id, invoice_id);
