ALTER TABLE transactions
    ADD COLUMN direction VARCHAR(7) NOT NULL DEFAULT 'EXPENSE',
    ADD COLUMN reference_month VARCHAR(7) NOT NULL DEFAULT '2026-01';

UPDATE transactions
SET reference_month = TO_CHAR(transaction_date, 'YYYY-MM')
WHERE reference_month = '2026-01';

CREATE INDEX idx_transactions_user_category_month
    ON transactions(user_id, category_id, reference_month)
    WHERE status = 'active' AND deleted_at IS NULL;
