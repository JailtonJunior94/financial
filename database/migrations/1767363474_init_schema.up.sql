-- =============================================================================
-- FINANCIAL SYSTEM - INITIAL SCHEMA
-- =============================================================================
-- Versão: 1.0.0
-- Compatível com: CockroachDB 22.x+ e PostgreSQL 14+
-- Data: 2026-01-02
--
-- Este schema unificado representa o estado inicial completo do banco de dados.
-- Todas as correções de precisão, tipos e relacionamentos foram aplicadas.
--
-- Decisões Arquiteturais:
-- - Valores monetários: NUMERIC(19,2) - máxima precisão e compatibilidade
-- - Timestamps: TIMESTAMPTZ - suporte a múltiplos fusos horários
-- - Soft Deletes: deleted_at em todas entidades de negócio
-- - Auditoria: created_at (NOT NULL) e updated_at (nullable)
-- - Isolamento: user_id em todas entidades principais
-- =============================================================================

-- =============================================================================
-- TABELA: users
-- Descrição: Usuários do sistema (root entity)
-- =============================================================================
CREATE TABLE users (
    id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(800) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uk_users_email UNIQUE (email)
);

-- Comentários para documentação
COMMENT ON TABLE users IS 'Usuários do sistema - Root Entity';
COMMENT ON COLUMN users.password IS 'Senha hasheada (bcrypt) - NUNCA armazenar em plain text';
COMMENT ON COLUMN users.email IS 'Email único do usuário - usado para autenticação';

-- =============================================================================
-- TABELA: categories
-- Descrição: Categorias financeiras hierárquicas por usuário
-- =============================================================================
CREATE TABLE categories (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    parent_id UUID,
    name VARCHAR(255) NOT NULL,
    sequence INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT fk_categories_users FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id)
        REFERENCES categories(id) ON DELETE RESTRICT
);

-- Comentários para documentação
COMMENT ON TABLE categories IS 'Categorias financeiras com suporte a hierarquia (parent-child)';
COMMENT ON COLUMN categories.parent_id IS 'FK para categoria pai - permite estrutura hierárquica';
COMMENT ON COLUMN categories.sequence IS 'Ordem de exibição dentro do mesmo nível hierárquico';

-- Índices otimizados para queries de categorias
CREATE INDEX idx_categories_list
    ON categories(user_id, sequence)
    WHERE parent_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idx_categories_user_active
    ON categories(user_id, deleted_at);

CREATE INDEX idx_categories_parent
    ON categories(parent_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_categories_user_parent
    ON categories(user_id, parent_id)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: payment_methods
-- Descrição: Métodos de pagamento globais (tabela de domínio)
-- =============================================================================
CREATE TABLE payment_methods (
    id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_payment_methods PRIMARY KEY (id),
    CONSTRAINT uk_payment_methods_code UNIQUE (code)
);

-- Comentários para documentação
COMMENT ON TABLE payment_methods IS 'Métodos de pagamento disponíveis no sistema (tabela global)';
COMMENT ON COLUMN payment_methods.code IS 'Código único do método (PIX, CREDIT_CARD, etc.)';

-- Índices
CREATE INDEX idx_payment_methods_code
    ON payment_methods(code)
    WHERE deleted_at IS NULL;

-- Seed inicial com métodos de pagamento brasileiros
INSERT INTO payment_methods (id, name, code, description, created_at) VALUES
    (gen_random_uuid(), 'PIX', 'PIX', 'Pagamento instantâneo via PIX', NOW()),
    (gen_random_uuid(), 'Cartão de Crédito', 'CREDIT_CARD', 'Pagamento com cartão de crédito', NOW()),
    (gen_random_uuid(), 'Cartão de Débito', 'DEBIT_CARD', 'Pagamento com cartão de débito', NOW()),
    (gen_random_uuid(), 'Boleto Bancário', 'BOLETO', 'Pagamento via boleto bancário', NOW()),
    (gen_random_uuid(), 'Dinheiro', 'CASH', 'Pagamento em dinheiro', NOW()),
    (gen_random_uuid(), 'Transferência Bancária', 'BANK_TRANSFER', 'Transferência bancária (TED/DOC)', NOW())
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- TABELA: cards
-- Descrição: Cartões de crédito dos usuários
-- =============================================================================
CREATE TABLE cards (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    due_day INT NOT NULL,
    closing_offset_days INT NOT NULL DEFAULT 7,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_cards PRIMARY KEY (id),
    CONSTRAINT fk_cards_users FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_cards_due_day
        CHECK (due_day >= 1 AND due_day <= 31),
    CONSTRAINT chk_cards_closing_offset_days
        CHECK (closing_offset_days >= 1 AND closing_offset_days <= 31)
);

-- Comentários para documentação
COMMENT ON TABLE cards IS 'Cartões de crédito dos usuários';
COMMENT ON COLUMN cards.due_day IS 'Dia do vencimento da fatura (1-31)';
COMMENT ON COLUMN cards.closing_offset_days IS 'Dias antes do vencimento para fechamento da fatura (padrão: 7)';

-- Índices
CREATE INDEX idx_cards_user_id
    ON cards(user_id)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: budgets
-- Descrição: Orçamentos mensais por usuário
-- =============================================================================
CREATE TABLE budgets (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    date DATE NOT NULL,
    amount_goal NUMERIC(19,2) NOT NULL,
    amount_used NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    percentage_used NUMERIC(6,3) NOT NULL DEFAULT 0.000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_budgets PRIMARY KEY (id),
    CONSTRAINT fk_budgets_users FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uk_budgets_user_date UNIQUE (user_id, date),
    CONSTRAINT chk_budgets_amount_goal
        CHECK (amount_goal > 0),
    CONSTRAINT chk_budgets_amount_used
        CHECK (amount_used >= 0),
    CONSTRAINT chk_budgets_percentage_used
        CHECK (percentage_used >= 0 AND percentage_used <= 100.000)
);

-- Comentários para documentação
COMMENT ON TABLE budgets IS 'Orçamentos mensais dos usuários (Aggregate Root)';
COMMENT ON COLUMN budgets.date IS 'Data de referência do orçamento (normalmente primeiro dia do mês)';
COMMENT ON COLUMN budgets.amount_goal IS 'Meta de valor total do orçamento';
COMMENT ON COLUMN budgets.amount_used IS 'Valor já utilizado (calculado)';
COMMENT ON COLUMN budgets.percentage_used IS 'Percentual utilizado (calculado)';

-- Índices
CREATE INDEX idx_budgets_user_id
    ON budgets(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_budgets_deleted_at
    ON budgets(deleted_at)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: budget_items
-- Descrição: Items individuais do orçamento por categoria
-- =============================================================================
CREATE TABLE budget_items (
    id UUID NOT NULL,
    budget_id UUID NOT NULL,
    category_id UUID NOT NULL,
    percentage_goal NUMERIC(6,3) NOT NULL DEFAULT 0.000,
    amount_goal NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    amount_used NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_budget_items PRIMARY KEY (id),
    CONSTRAINT fk_budget_items_budgets FOREIGN KEY (budget_id)
        REFERENCES budgets(id) ON DELETE CASCADE,
    CONSTRAINT fk_budget_items_categories FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE RESTRICT,
    CONSTRAINT uk_budget_items_budget_category
        UNIQUE (budget_id, category_id),
    CONSTRAINT chk_budget_items_percentage_goal
        CHECK (percentage_goal >= 0 AND percentage_goal <= 100.000),
    CONSTRAINT chk_budget_items_amount_goal
        CHECK (amount_goal >= 0),
    CONSTRAINT chk_budget_items_amount_used
        CHECK (amount_used >= 0)
);

-- Comentários para documentação
COMMENT ON TABLE budget_items IS 'Items do orçamento distribuídos por categoria';
COMMENT ON COLUMN budget_items.percentage_goal IS 'Percentual do orçamento total alocado para esta categoria';
COMMENT ON COLUMN budget_items.amount_goal IS 'Valor calculado baseado no percentage_goal';
COMMENT ON COLUMN budget_items.amount_used IS 'Valor já gasto nesta categoria';

-- Índices
CREATE INDEX idx_budget_items_budget_id
    ON budget_items(budget_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_budget_items_deleted_at
    ON budget_items(deleted_at)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: invoices
-- Descrição: Faturas mensais de cartões de crédito
-- =============================================================================
CREATE TABLE invoices (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    card_id UUID NOT NULL,
    reference_month DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_invoices PRIMARY KEY (id),
    CONSTRAINT fk_invoices_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_invoices_card FOREIGN KEY (card_id)
        REFERENCES cards(id) ON DELETE RESTRICT,
    CONSTRAINT uk_invoices_user_card_month
        UNIQUE (user_id, card_id, reference_month),
    CONSTRAINT chk_invoices_total_amount
        CHECK (total_amount >= 0)
);

-- Comentários para documentação
COMMENT ON TABLE invoices IS 'Faturas mensais de cartões de crédito (Aggregate Root)';
COMMENT ON COLUMN invoices.reference_month IS 'Mês de referência da fatura (formato DATE primeiro dia do mês)';
COMMENT ON COLUMN invoices.due_date IS 'Data de vencimento da fatura';
COMMENT ON COLUMN invoices.total_amount IS 'Valor total da fatura (soma dos invoice_items)';

-- Índices otimizados
CREATE INDEX idx_invoices_user_id
    ON invoices(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoices_card_id
    ON invoices(card_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoices_reference_month
    ON invoices(reference_month)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoices_user_reference_month
    ON invoices(user_id, reference_month)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoices_due_date
    ON invoices(due_date)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: invoice_items
-- Descrição: Items individuais (compras) de uma fatura
-- =============================================================================
CREATE TABLE invoice_items (
    id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    category_id UUID NOT NULL,
    purchase_date DATE NOT NULL,
    description VARCHAR(255) NOT NULL,
    total_amount NUMERIC(19,2) NOT NULL,
    installment_number INT NOT NULL,
    installment_total INT NOT NULL,
    installment_amount NUMERIC(19,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_invoice_items PRIMARY KEY (id),
    CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id)
        REFERENCES invoices(id) ON DELETE CASCADE,
    CONSTRAINT fk_invoice_items_category FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE RESTRICT,
    CONSTRAINT chk_invoice_items_installment_number
        CHECK (installment_number > 0 AND installment_number <= installment_total),
    CONSTRAINT chk_invoice_items_installment_total
        CHECK (installment_total > 0),
    CONSTRAINT chk_invoice_items_amounts
        CHECK (total_amount > 0 AND installment_amount > 0)
);

-- Comentários para documentação
COMMENT ON TABLE invoice_items IS 'Compras individuais que compõem uma fatura';
COMMENT ON COLUMN invoice_items.total_amount IS 'Valor total original da compra';
COMMENT ON COLUMN invoice_items.installment_number IS 'Número da parcela atual (1-N)';
COMMENT ON COLUMN invoice_items.installment_total IS 'Total de parcelas da compra';
COMMENT ON COLUMN invoice_items.installment_amount IS 'Valor desta parcela específica';

-- Índices otimizados
CREATE INDEX idx_invoice_items_invoice_id
    ON invoice_items(invoice_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoice_items_category_id
    ON invoice_items(category_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoice_items_purchase_date
    ON invoice_items(purchase_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoice_items_purchase_origin
    ON invoice_items(purchase_date, category_id, description)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_invoice_items_invoice_purchase
    ON invoice_items(invoice_id, purchase_date, installment_number)
    WHERE deleted_at IS NULL;

-- =============================================================================
-- TABELA: monthly_transactions
-- Descrição: Consolidado financeiro mensal do usuário
-- =============================================================================
CREATE TABLE monthly_transactions (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    reference_month VARCHAR(7) NOT NULL,
    total_income NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    total_expense NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    total_amount NUMERIC(19,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,

    CONSTRAINT pk_monthly_transactions PRIMARY KEY (id),
    CONSTRAINT fk_monthly_transactions_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uk_monthly_transactions_user_month
        UNIQUE (user_id, reference_month),
    CONSTRAINT chk_monthly_transactions_totals
        CHECK (total_income >= 0 AND total_expense >= 0)
);

-- Comentários para documentação
COMMENT ON TABLE monthly_transactions IS 'Consolidado financeiro mensal do usuário (Aggregate Root)';
COMMENT ON COLUMN monthly_transactions.reference_month IS 'Mês de referência no formato YYYY-MM';
COMMENT ON COLUMN monthly_transactions.total_income IS 'Total de receitas do mês';
COMMENT ON COLUMN monthly_transactions.total_expense IS 'Total de despesas do mês';
COMMENT ON COLUMN monthly_transactions.total_amount IS 'Saldo do mês (income - expense)';

-- Índices
CREATE INDEX idx_monthly_transactions_user
    ON monthly_transactions(user_id);

CREATE INDEX idx_monthly_transactions_reference
    ON monthly_transactions(reference_month);

CREATE INDEX idx_monthly_transactions_user_month
    ON monthly_transactions(user_id, reference_month);

-- =============================================================================
-- TABELA: transaction_items
-- Descrição: Transações individuais do mês
-- =============================================================================
CREATE TABLE transaction_items (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    monthly_id UUID NOT NULL,
    category_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount NUMERIC(19,2) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_transaction_items PRIMARY KEY (id),
    CONSTRAINT fk_transaction_items_monthly_transaction FOREIGN KEY (monthly_id)
        REFERENCES monthly_transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_transaction_items_category FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE RESTRICT,
    CONSTRAINT chk_transaction_items_direction
        CHECK (direction IN ('INCOME', 'EXPENSE')),
    CONSTRAINT chk_transaction_items_type
        CHECK (type IN ('PIX', 'BOLETO', 'TRANSFER', 'CREDIT_CARD')),
    CONSTRAINT chk_transaction_items_amount
        CHECK (amount >= 0)
);

-- Comentários para documentação
COMMENT ON TABLE transaction_items IS 'Transações individuais que compõem o consolidado mensal';
COMMENT ON COLUMN transaction_items.monthly_id IS 'FK para o aggregate root (MonthlyTransaction)';
COMMENT ON COLUMN transaction_items.direction IS 'Direção da transação: INCOME (entrada) ou EXPENSE (saída)';
COMMENT ON COLUMN transaction_items.type IS 'Tipo de pagamento: PIX, BOLETO, TRANSFER, CREDIT_CARD';
COMMENT ON COLUMN transaction_items.is_paid IS 'Indica se a transação foi efetivada/paga';
COMMENT ON COLUMN transaction_items.deleted_at IS 'Soft delete - transações deletadas são ignoradas nos cálculos';

-- Índices otimizados
CREATE INDEX idx_transaction_items_monthly
    ON transaction_items(monthly_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_transaction_items_category
    ON transaction_items(category_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_transaction_items_type
    ON transaction_items(type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_transaction_items_deleted
    ON transaction_items(deleted_at);

-- =============================================================================
-- TABELA: outbox_events
-- Descrição: Outbox Pattern - eventos de domínio para publicação
-- =============================================================================
CREATE TABLE outbox_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_outbox_events PRIMARY KEY (id),
    CONSTRAINT chk_outbox_events_status
        CHECK (status IN ('pending', 'published', 'failed')),
    CONSTRAINT chk_outbox_events_retry_count
        CHECK (retry_count >= 0 AND retry_count <= 3)
);

-- Comentários para documentação
COMMENT ON TABLE outbox_events IS 'Armazena eventos de domínio para publicação confiável (Outbox Pattern)';
COMMENT ON COLUMN outbox_events.status IS 'Status do evento: pending, published, failed';
COMMENT ON COLUMN outbox_events.retry_count IS 'Número de tentativas de publicação (máximo 3)';
COMMENT ON COLUMN outbox_events.payload IS 'Payload do evento em formato JSON';

-- Índices otimizados
CREATE INDEX idx_outbox_status_created
    ON outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_aggregate
    ON outbox_events(aggregate_type, aggregate_id);

CREATE INDEX idx_outbox_created_at
    ON outbox_events(created_at DESC);

-- =============================================================================
-- TABELA: processed_events
-- Descrição: Rastreamento de eventos processados (idempotência)
-- =============================================================================
CREATE TABLE processed_events (
    event_id UUID NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_processed_events PRIMARY KEY (event_id, consumer_name)
);

-- Comentários para documentação
COMMENT ON TABLE processed_events IS 'Rastreia eventos já processados para garantir idempotência dos consumers';
COMMENT ON COLUMN processed_events.event_id IS 'ID do evento da tabela outbox_events';
COMMENT ON COLUMN processed_events.consumer_name IS 'Nome do consumer que processou o evento';
COMMENT ON COLUMN processed_events.processed_at IS 'Timestamp de quando o evento foi processado com sucesso';

-- Índices
CREATE INDEX idx_processed_events_processed_at
    ON processed_events(processed_at DESC);

-- =============================================================================
-- FIM DO SCHEMA INICIAL
-- =============================================================================
