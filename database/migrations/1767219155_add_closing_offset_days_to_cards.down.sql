-- Remove campo closing_offset_days da tabela cards

ALTER TABLE cards
DROP CONSTRAINT IF EXISTS chk_cards_closing_offset_days;

ALTER TABLE cards
DROP COLUMN IF EXISTS closing_offset_days;
