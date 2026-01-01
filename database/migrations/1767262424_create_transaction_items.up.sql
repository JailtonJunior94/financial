-- Tabela de itens de transação
CREATE TABLE IF NOT EXISTS transaction_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monthly_id UUID NOT NULL REFERENCES monthly_transactions(id) ON DELETE CASCADE,
    category_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount NUMERIC(19, 2) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_transaction_items_direction CHECK (direction IN ('INCOME', 'EXPENSE')),
    CONSTRAINT chk_transaction_items_type CHECK (type IN ('PIX', 'BOLETO', 'TRANSFER', 'CREDIT_CARD')),
    CONSTRAINT chk_transaction_items_amount CHECK (amount >= 0)
);

-- Índices para performance
CREATE INDEX idx_transaction_items_monthly ON transaction_items(monthly_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transaction_items_category ON transaction_items(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transaction_items_type ON transaction_items(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_transaction_items_deleted ON transaction_items(deleted_at);

-- Comentários
COMMENT ON TABLE transaction_items IS 'Itens individuais de transação (pertence ao aggregate MonthlyTransaction)';
COMMENT ON COLUMN transaction_items.monthly_id IS 'FK para o aggregate root (MonthlyTransaction)';
COMMENT ON COLUMN transaction_items.direction IS 'Direção: INCOME (entrada) ou EXPENSE (saída)';
COMMENT ON COLUMN transaction_items.type IS 'Tipo: PIX, BOLETO, TRANSFER, CREDIT_CARD';
COMMENT ON COLUMN transaction_items.is_paid IS 'Indica se a transação foi efetivada';
COMMENT ON COLUMN transaction_items.deleted_at IS 'Soft delete - items deletados são ignorados nos cálculos';
