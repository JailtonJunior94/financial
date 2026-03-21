DROP INDEX IF EXISTS idx_transactions_user_category_month;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS reference_month,
    DROP COLUMN IF EXISTS direction;
