-- Migration 002: Compliance Core (Idempotence, Audit Chain, Outbox)

CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id       TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash    TEXT NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('PENDING', 'PROCESSED', 'FAILED')),
    response_code   INTEGER,
    response_body   TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (tenant_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS invoices (
    id              TEXT NOT NULL,
    tenant_id       TEXT NOT NULL,
    invoice_number  TEXT NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('DRAFT', 'ISSUED', 'TRANSMITTED', 'REJECTED')),
    document_format TEXT NOT NULL CHECK (document_format IN ('UBL_2_1', 'CII_D16B', 'FACTUR_X_BASIC', 'FACTUR_X_COMFORT')),
    payload_hash    TEXT NOT NULL,
    raw_payload     BYTEA NOT NULL,
    net_amount      NUMERIC(15, 2) NOT NULL,
    tax_amount      NUMERIC(15, 2) NOT NULL,
    total_amount    NUMERIC(15, 2) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, invoice_number)
);

CREATE TABLE IF NOT EXISTS audit_events (
    id                  TEXT PRIMARY KEY,
    tenant_id           TEXT NOT NULL,
    invoice_id          TEXT NOT NULL,
    sequence_num        BIGINT NOT NULL,
    event_type          TEXT NOT NULL,
    payload_hash        TEXT NOT NULL,
    prev_event_hash     TEXT NOT NULL,
    event_hash          TEXT NOT NULL,
    metadata            JSONB DEFAULT '{}'::jsonb,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, invoice_id, sequence_num),
    FOREIGN KEY (tenant_id, invoice_id) REFERENCES invoices(tenant_id, id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS transactional_outbox (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    invoice_id      TEXT NOT NULL,
    destination_id  TEXT NOT NULL,
    transport_type  TEXT NOT NULL DEFAULT 'AS4',
    payload_ref     TEXT NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'DELIVERED', 'FAILED', 'DEAD_LETTER')),
    retry_count     INTEGER DEFAULT 0,
    max_retries     INTEGER DEFAULT 5,
    last_error      TEXT,
    next_retry_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_outbox_fetch ON transactional_outbox (status, next_retry_at)
WHERE status IN ('PENDING', 'FAILED');