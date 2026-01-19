package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// CardBillingInfo contém informações de faturamento do cartão
// necessárias para calcular a fatura correta.
// ✅ Fonte da verdade: módulo cards.
type CardBillingInfo struct {
	CardID            vos.UUID
	DueDay            int // Dia de vencimento da fatura (1-31)
	ClosingOffsetDays int // Quantos dias ANTES do vencimento fecha a fatura (padrão brasileiro: 7)
}

// CardProvider é uma porta de domínio que abstrai o acesso a dados do cartão.
// Implementação deve ficar na infraestrutura do módulo cards.
type CardProvider interface {
	// GetCardBillingInfo obtém informações de faturamento do cartão
	// Valida que o cartão pertence ao usuário
	GetCardBillingInfo(ctx context.Context, userID vos.UUID, cardID vos.UUID) (*CardBillingInfo, error)
}
