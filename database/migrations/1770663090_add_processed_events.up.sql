-- Migration: Add processed_events table for consumer idempotency
-- Purpose: Track processed outbox events to prevent duplicate processing

CREATE TABLE processed_events (
    event_id UUID NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_processed_events PRIMARY KEY (event_id, consumer_name)
);

-- Índice para cleanup de eventos antigos
CREATE INDEX idx_processed_events_processed_at
    ON processed_events(processed_at);

-- Comentários
COMMENT ON TABLE processed_events IS 'Rastreia eventos já processados pelos consumers para garantir idempotência';
COMMENT ON COLUMN processed_events.event_id IS 'ID do evento do outbox_events';
COMMENT ON COLUMN processed_events.consumer_name IS 'Nome do consumer que processou o evento';
COMMENT ON COLUMN processed_events.processed_at IS 'Timestamp de quando o evento foi processado';
