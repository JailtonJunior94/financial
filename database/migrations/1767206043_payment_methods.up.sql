CREATE TABLE payment_methods (
    id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_payment_methods PRIMARY KEY (id),
    CONSTRAINT uk_payment_methods_code UNIQUE (code)
);

CREATE INDEX idx_payment_methods_deleted_at ON payment_methods(deleted_at);
CREATE INDEX idx_payment_methods_code ON payment_methods(code) WHERE deleted_at IS NULL;

-- Seed inicial com as principais formas de pagamento brasileiras
INSERT INTO payment_methods (id, name, code, description, created_at) VALUES
    (gen_random_uuid(), 'PIX', 'PIX', 'Pagamento instantâneo via PIX', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'Cartão de Crédito', 'CREDIT_CARD', 'Pagamento com cartão de crédito', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'Cartão de Débito', 'DEBIT_CARD', 'Pagamento com cartão de débito', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'Boleto Bancário', 'BOLETO', 'Pagamento via boleto bancário', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'Dinheiro', 'CASH', 'Pagamento em dinheiro', CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'Transferência Bancária', 'BANK_TRANSFER', 'Transferência bancária (TED/DOC)', CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;
