DROP INDEX IF EXISTS idx_outbox_pending_retry;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS next_retry_at;
