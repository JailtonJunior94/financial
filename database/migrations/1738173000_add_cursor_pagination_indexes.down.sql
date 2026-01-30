-- Migration Down: Remove cursor-based pagination indexes
-- Reverte a migration 1738173000_add_cursor_pagination_indexes.up.sql

DROP INDEX CONCURRENTLY IF EXISTS idx_cards_user_name_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_categories_user_seq_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_user_month_due_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_invoices_card_month_id;
