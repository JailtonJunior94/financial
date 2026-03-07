DROP INDEX IF EXISTS idx_invoices_user_status;
DROP INDEX IF EXISTS idx_invoices_card_month;
ALTER TABLE invoices DROP COLUMN IF EXISTS closing_date;
ALTER TABLE invoices DROP COLUMN IF EXISTS status;
