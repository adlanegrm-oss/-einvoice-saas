-- Migration 002: Persistent Replay & Outbox Engine
CREATE TABLE IF NOT EXISTS inbound_messages (
    tenant_id VARCHAR(64) NOT NULL,
    message_id VARCHAR(128) NOT NULL,
    protocol VARCHAR(32) NOT NULL DEFAULT 'AS4',
    payload_digest CHAR(64) NOT NULL,
    sender_id VARCHAR(128) NOT NULL,
    received_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_inbound_digest ON inbound_messages(payload_digest);

CREATE TABLE IF NOT EXISTS webhook_events (
    tenant_id VARCHAR(64) NOT NULL,
    event_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    signature_hash VARCHAR(128) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, event_id)
);

CREATE TABLE IF NOT EXISTS outbox_dispatches (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    invoice_number VARCHAR(128) NOT NULL,
    destination_pdp VARCHAR(64) NOT NULL,
    format VARCHAR(32) NOT NULL,
    payload_xml TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
