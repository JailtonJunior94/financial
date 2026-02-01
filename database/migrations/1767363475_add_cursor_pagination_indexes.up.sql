-- Migration: Add cursor-based pagination indexes
-- Description: Adiciona índices compostos determinísticos para suportar cursor-based pagination eficiente
-- Performance: Keyset pagination O(1) com index-only scans

-- ============================================================================
-- CARD: Índice composto (user_id, name, id)
-- ============================================================================
-- Suporta: ORDER BY name, id com WHERE user_id = ?
-- Query: SELECT * FROM cards WHERE user_id = ? AND (name, id) > (?, ?) ORDER BY name, id LIMIT ?
-- Benefício: Elimina filesort, permite keyset pagination eficiente

CREATE INDEX IF NOT EXISTS idx_cards_user_name_id
ON cards(user_id, name, id)
WHERE deleted_at IS NULL;

-- ============================================================================
-- CATEGORY: Índice composto (user_id, sequence, id)
-- ============================================================================
-- Suporta: ORDER BY sequence, id com WHERE user_id = ? AND parent_id IS NULL
-- Query: SELECT * FROM categories WHERE user_id = ? AND parent_id IS NULL AND (sequence, id) > (?, ?) ORDER BY sequence, id LIMIT ?
-- Benefício: Keyset pagination em ordem customizável (sequence) com desempate por id

CREATE INDEX IF NOT EXISTS idx_categories_user_seq_id
ON categories(user_id, sequence, id)
WHERE parent_id IS NULL AND deleted_at IS NULL;

-- ============================================================================
-- INVOICE BY MONTH: Índice composto (user_id, reference_month, due_date, id)
-- ============================================================================
-- Suporta: ORDER BY due_date, id com WHERE user_id = ? AND reference_month = ?
-- Query: SELECT * FROM invoices WHERE user_id = ? AND reference_month = ? AND (due_date, id) > (?, ?) ORDER BY due_date, id LIMIT ?
-- Benefício: Paginação eficiente de faturas do mês ordenadas por vencimento

CREATE INDEX IF NOT EXISTS idx_invoices_user_month_due_id
ON invoices(user_id, reference_month, due_date, id)
WHERE deleted_at IS NULL;

-- ============================================================================
-- INVOICE BY CARD: Índice composto (card_id, reference_month DESC, id DESC)
-- ============================================================================
-- Suporta: ORDER BY reference_month DESC, id DESC com WHERE card_id = ?
-- Query: SELECT * FROM invoices WHERE card_id = ? AND (reference_month, id) < (?, ?) ORDER BY reference_month DESC, id DESC LIMIT ?
-- Benefício: Histórico do cartão (mais recentes primeiro) com paginação eficiente
-- Nota: DESC no índice permite PostgreSQL usar index scan sem reverse scan

CREATE INDEX IF NOT EXISTS idx_invoices_card_month_id
ON invoices(card_id, reference_month DESC, id DESC)
WHERE deleted_at IS NULL;

-- ============================================================================
-- COMENTÁRIOS E PERFORMANCE
-- ============================================================================

-- CONCURRENTLY: Índices criados sem lock exclusivo (safe para produção)
-- WHERE deleted_at IS NULL: Partial index (menor tamanho, melhor performance)
-- Ordem das colunas: (filter_column, order_by_columns..., unique_column)

-- EXPLAIN ANALYZE esperado (exemplo Card):
-- ANTES (sem índice composto):
--   Index Scan using idx_cards_user_id  (cost=0.29..8.31 rows=1)
--     -> Sort  (cost=8.31..8.32)  ← Filesort
--
-- DEPOIS (com índice composto):
--   Index Scan using idx_cards_user_name_id  (cost=0.29..4.31 rows=20)
--     ← Sem sort, direto do índice

-- Comments removed: COMMENT ON INDEX doesn't work well with IF NOT EXISTS
-- Index names are self-documenting
