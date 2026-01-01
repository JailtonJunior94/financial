package adapters

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

// CardProviderAdapter implementa a interface CardProvider do módulo invoice.
// Permite que invoice obtenha dados de faturamento sem acoplamento direto.
type CardProviderAdapter struct {
	cardRepository interfaces.CardRepository
	o11y           observability.Observability
}

// NewCardProviderAdapter cria um novo adapter de provedor de cartão.
func NewCardProviderAdapter(
	cardRepository interfaces.CardRepository,
	o11y observability.Observability,
) invoiceInterfaces.CardProvider {
	return &CardProviderAdapter{
		cardRepository: cardRepository,
		o11y:           o11y,
	}
}

// GetCardBillingInfo obtém informações de faturamento do cartão.
// Valida que o cartão pertence ao usuário.
func (a *CardProviderAdapter) GetCardBillingInfo(
	ctx context.Context,
	userID vos.UUID,
	cardID vos.UUID,
) (*invoiceInterfaces.CardBillingInfo, error) {
	ctx, span := a.o11y.Tracer().Start(ctx, "card_provider_adapter.get_card_billing_info")
	defer span.End()

	// Busca o cartão (já validando ownership internamente)
	card, err := a.cardRepository.FindByID(ctx, userID, cardID)
	if err != nil {
		a.o11y.Logger().Error(ctx, "failed to find card", observability.Error(err))
		return nil, err
	}

	if card == nil {
		return nil, customErrors.ErrCardNotFound
	}

	// Retorna apenas as informações necessárias para faturamento
	// ✅ Cards é a fonte da verdade sobre ciclo de faturamento
	return &invoiceInterfaces.CardBillingInfo{
		CardID:            card.ID,
		DueDay:            card.DueDay.Value,
		ClosingOffsetDays: card.ClosingOffsetDays.Value,
	}, nil
}
