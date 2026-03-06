-- Remover indice
DROP INDEX IF EXISTS idx_cards_user_id_type;

-- Remover CHECK constraints novas
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_type;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_flag;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_last_four_digits;

-- Remover CHECK constraints recriadas
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_due_day;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS chk_cards_closing_offset_days;

-- Restaurar NOT NULL
ALTER TABLE cards ALTER COLUMN due_day SET NOT NULL;
ALTER TABLE cards ALTER COLUMN closing_offset_days SET NOT NULL;

-- Restaurar CHECK constraints originais
ALTER TABLE cards ADD CONSTRAINT chk_cards_due_day
    CHECK (due_day >= 1 AND due_day <= 31);
ALTER TABLE cards ADD CONSTRAINT chk_cards_closing_offset_days
    CHECK (closing_offset_days >= 1 AND closing_offset_days <= 31);

-- Remover colunas novas
ALTER TABLE cards DROP COLUMN IF EXISTS type;
ALTER TABLE cards DROP COLUMN IF EXISTS flag;
ALTER TABLE cards DROP COLUMN IF EXISTS last_four_digits;
