-- Table de déduplication et protection anti-rejeu partagée (AS4, Webhooks)
CREATE TABLE IF NOT EXISTS processed_messages (
    tenant_id    VARCHAR(64)  NOT NULL,
    source_type  VARCHAR(32)  NOT NULL, -- 'AS4', 'WEBHOOK', 'PDP'
    message_id   VARCHAR(255) NOT NULL,
    payload_hash CHAR(64)     NOT NULL,
    processed_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, source_type, message_id)
);
CREATE INDEX IF NOT EXISTS idx_processed_messages_date ON processed_messages (processed_at);