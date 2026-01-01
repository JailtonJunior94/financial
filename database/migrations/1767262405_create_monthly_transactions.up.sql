-- Tabela de consolidado mensal de transações
CREATE TABLE IF NOT EXISTS monthly_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    reference_month VARCHAR(7) NOT NULL, -- YYYY-MM
    total_income NUMERIC(19, 2) NOT NULL DEFAULT 0.00,
    total_expense NUMERIC(19, 2) NOT NULL DEFAULT 0.00,
    total_amount NUMERIC(19, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    
    CONSTRAINT uq_monthly_transactions_user_month UNIQUE(user_id, reference_month)
);

-- Índices para performance
CREATE INDEX idx_monthly_transactions_user ON monthly_transactions(user_id);
CREATE INDEX idx_monthly_transactions_reference ON monthly_transactions(reference_month);

-- Comentários
COMMENT ON TABLE monthly_transactions IS 'Consolidado financeiro mensal do usuário (Aggregate Root)';
COMMENT ON COLUMN monthly_transactions.reference_month IS 'Mês de referência no formato YYYY-MM';
COMMENT ON COLUMN monthly_transactions.total_income IS 'Total de receitas do mês';
COMMENT ON COLUMN monthly_transactions.total_expense IS 'Total de despesas do mês';
COMMENT ON COLUMN monthly_transactions.total_amount IS 'Saldo do mês (income - expense)';
