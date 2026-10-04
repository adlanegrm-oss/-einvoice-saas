-- Table d'idempotence des requêtes API
CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id        VARCHAR(64)  NOT NULL,
    idempotency_key  VARCHAR(255) NOT NULL,
    invoice_id       VARCHAR(64)  NOT NULL,
    request_hash     CHAR(64)     NOT NULL,
    response_code    INT          NOT NULL,
    response_body    TEXT         NOT NULL,
    created_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, idempotency_key)
);

-- Factures avec isolation multi-tenant et statut
CREATE TABLE IF NOT EXISTS invoices (
    id               VARCHAR(64)  NOT NULL,
    tenant_id        VARCHAR(64)  NOT NULL,
    invoice_number   VARCHAR(128) NOT NULL,
    status           VARCHAR(32)  NOT NULL,
    total_ht         BIGINT       NOT NULL,
    total_tva        BIGINT       NOT NULL,
    total_ttc        BIGINT       NOT NULL,
    document_hash    CHAR(64)     NOT NULL,
    created_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, invoice_number)
);

-- Outbox transactionnelle persistante
CREATE TABLE IF NOT EXISTS outbox_events (
    id               VARCHAR(64)  PRIMARY KEY,
    tenant_id        VARCHAR(64)  NOT NULL,
    aggregate_id     VARCHAR(64)  NOT NULL,
    event_type       VARCHAR(64)  NOT NULL,
    payload          TEXT         NOT NULL,
    status           VARCHAR(32)  NOT NULL DEFAULT 'PENDING',
    attempts         INT          NOT NULL DEFAULT 0,
    max_attempts     INT          NOT NULL DEFAULT 5,
    next_attempt_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    created_at       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    processed_at     TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_outbox_queue 
ON outbox_events (status, next_attempt_at);