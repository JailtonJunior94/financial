-- Adiciona campo closing_offset_days à tabela cards
-- Representa quantos dias ANTES do vencimento ocorre o fechamento da fatura
-- Padrão brasileiro (Nubank, BTG, XP): 7 dias

ALTER TABLE cards
ADD COLUMN closing_offset_days INT NOT NULL DEFAULT 7;

-- Adiciona constraint para garantir valor válido
ALTER TABLE cards
ADD CONSTRAINT chk_cards_closing_offset_days
CHECK (closing_offset_days >= 1 AND closing_offset_days <= 31);

-- Comentários para documentação
COMMENT ON COLUMN cards.closing_offset_days IS 'Quantos dias antes do vencimento ocorre o fechamento da fatura (padrão brasileiro: 7)';
