DROP INDEX IF EXISTS idx_transaction_items_is_paid;
DROP INDEX IF EXISTS idx_transaction_items_type;
DROP INDEX IF EXISTS idx_transaction_items_direction;
DROP INDEX IF EXISTS idx_transaction_items_deleted_at;
DROP INDEX IF EXISTS idx_transaction_items_category_id;
DROP INDEX IF EXISTS idx_transaction_items_monthly_transaction_id;

DROP TABLE IF EXISTS transaction_items;
