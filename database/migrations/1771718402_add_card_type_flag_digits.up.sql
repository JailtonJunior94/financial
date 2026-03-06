-- Novos campos
ALTER TABLE cards ADD COLUMN IF NOT EXISTS type VARCHAR(10) NOT NULL DEFAULT 'credit';
ALTER TABLE cards ADD COLUMN IF NOT EXISTS flag VARCHAR(20) NOT NULL DEFAULT 'visa';
ALTER TABLE cards ADD COLUMN IF NOT EXISTS last_four_digits VARCHAR(4) NOT NULL DEFAULT '0000';

-- Permitir NULL para debito
ALTER TABLE cards ALTER COLUMN due_day DROP NOT NULL;
ALTER TABLE cards ALTER COLUMN closing_offset_days DROP NOT NULL;

-- Remover CHECK constraints existentes (nao aceitam NULL)
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_due_day;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_closing_offset_days;

-- Recriar CHECK constraints permitindo NULL
ALTER TABLE cards ADD CONSTRAINT chk_cards_due_day
    CHECK (due_day IS NULL OR (due_day >= 1 AND due_day <= 31));
ALTER TABLE cards ADD CONSTRAINT chk_cards_closing_offset_days
    CHECK (closing_offset_days IS NULL OR (closing_offset_days >= 1 AND closing_offset_days <= 31));

-- CHECK para type e flag
ALTER TABLE cards ADD CONSTRAINT chk_cards_type
    CHECK (type IN ('credit', 'debit'));
ALTER TABLE cards ADD CONSTRAINT chk_cards_flag
    CHECK (flag IN ('visa', 'mastercard', 'elo', 'amex', 'hipercard'));

-- CHECK: last_four_digits deve ter exatamente 4 digitos numericos
ALTER TABLE cards ADD CONSTRAINT chk_cards_last_four_digits
    CHECK (last_four_digits ~ '^[0-9]{4}$');

-- Indice para busca por tipo
CREATE INDEX IF NOT EXISTS idx_cards_user_id_type ON cards(user_id, type) WHERE deleted_at IS NULL;
