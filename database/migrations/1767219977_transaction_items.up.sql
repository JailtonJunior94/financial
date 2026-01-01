CREATE TABLE IF NOT EXISTS transaction_items (
    id UUID PRIMARY KEY,
    monthly_transaction_id UUID NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('INCOME', 'EXPENSE')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('PIX', 'BOLETO', 'TRANSFER', 'CREDIT_CARD')),
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    category_id UUID NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_transaction_items_monthly_transaction FOREIGN KEY (monthly_transaction_id) REFERENCES monthly_transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_transaction_items_category FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE INDEX idx_transaction_items_monthly_transaction_id ON transaction_items(monthly_transaction_id);
CREATE INDEX idx_transaction_items_category_id ON transaction_items(category_id);
CREATE INDEX idx_transaction_items_deleted_at ON transaction_items(deleted_at);
CREATE INDEX idx_transaction_items_direction ON transaction_items(direction);
CREATE INDEX idx_transaction_items_type ON transaction_items(type);
CREATE INDEX idx_transaction_items_is_paid ON transaction_items(is_paid);
