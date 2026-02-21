-- Migration: Add next_retry_at to outbox_events for exponential backoff
-- Purpose: Allow the dispatcher to skip events that are not yet ready for retry,
--          implementing true exponential backoff instead of retrying on every poll cycle.
--
-- Backoff schedule (set by the application on IncrementRetry):
--   retry_count=1 → next_retry_at = now + 30s
--   retry_count=2 → next_retry_at = now + 2m
--
-- The FindPendingBatch query filters: next_retry_at IS NULL OR next_retry_at <= NOW()
-- New events (retry_count=0) always have next_retry_at = NULL → processed immediately.

ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

-- Partial index: only pending events need this filter
CREATE INDEX IF NOT EXISTS idx_outbox_pending_retry
    ON outbox_events(status, next_retry_at)
    WHERE status = 'pending';
