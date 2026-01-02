-- =============================================================================
-- FINANCIAL SYSTEM - ROLLBACK INITIAL SCHEMA
-- =============================================================================
-- Este arquivo reverte completamente o schema inicial
-- Ordem: reversa à criação (devido às dependências de FKs)
-- =============================================================================

-- Tabelas sem dependências externas (podem ser removidas por último)
DROP TABLE IF EXISTS processed_events CASCADE;
DROP TABLE IF EXISTS outbox_events CASCADE;

-- Transaction items e monthly transactions (relacionados)
DROP TABLE IF EXISTS transaction_items CASCADE;
DROP TABLE IF EXISTS monthly_transactions CASCADE;

-- Invoice items e invoices (relacionados)
DROP TABLE IF EXISTS invoice_items CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;

-- Budget items e budgets (relacionados)
DROP TABLE IF EXISTS budget_items CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;

-- Cards
DROP TABLE IF EXISTS cards CASCADE;

-- Payment methods (tabela global)
DROP TABLE IF EXISTS payment_methods CASCADE;

-- Categories (hierárquico - tem FK para si mesmo)
DROP TABLE IF EXISTS categories CASCADE;

-- Users (root entity - removida por último)
DROP TABLE IF EXISTS users CASCADE;

-- =============================================================================
-- FIM DO ROLLBACK
-- =============================================================================
