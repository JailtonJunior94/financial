ALTER TABLE invoices ADD COLUMN IF NOT EXISTS status VARCHAR(10) NOT NULL DEFAULT 'open'
    CHECK (status IN ('open','closed','paid'));

ALTER TABLE invoices ADD COLUMN IF NOT EXISTS closing_date DATE;

CREATE INDEX IF NOT EXISTS idx_invoices_card_month
    ON invoices(card_id, reference_month);

CREATE INDEX IF NOT EXISTS idx_invoices_user_status
    ON invoices(user_id, status);
