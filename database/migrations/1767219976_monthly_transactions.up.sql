CREATE TABLE IF NOT EXISTS monthly_transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    reference_month VARCHAR(7) NOT NULL,
    total_income NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total_expense NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NULL,
    
    CONSTRAINT fk_monthly_transactions_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uk_monthly_transactions_user_month UNIQUE (user_id, reference_month)
);

CREATE INDEX idx_monthly_transactions_user_id ON monthly_transactions(user_id);
CREATE INDEX idx_monthly_transactions_reference_month ON monthly_transactions(reference_month);
CREATE INDEX idx_monthly_transactions_user_month ON monthly_transactions(user_id, reference_month);
