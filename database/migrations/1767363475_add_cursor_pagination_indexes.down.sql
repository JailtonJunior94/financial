-- Migration Down: Remove cursor-based pagination indexes
-- Reverte a migration 1767363475_add_cursor_pagination_indexes.up.sql
--
-- NOTA: CONCURRENTLY removido para garantir compatibilidade com PostgreSQL
-- (DROP INDEX CONCURRENTLY não pode ser executado dentro de transação).
-- Em CockroachDB o comportamento é equivalente pois DDL é sempre transacional.

DROP INDEX IF EXISTS idx_cards_user_name_id;
DROP INDEX IF EXISTS idx_categories_user_seq_id;
DROP INDEX IF EXISTS idx_invoices_user_month_due_id;
DROP INDEX IF EXISTS idx_invoices_card_month_id;
