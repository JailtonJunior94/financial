package card

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/card/application/usecase"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type CardModule struct {
	CardRouter   *http.CardRouter
	CardProvider invoiceInterfaces.CardProvider
}

func NewCardModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) CardModule {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	// Inicializa métricas do módulo de cartões
	cardMetrics := metrics.NewCardMetrics(o11y)

	cardRepository := repositories.NewCardRepository(db, o11y)
	findCardUsecase := usecase.NewFindCardUseCase(o11y, cardRepository, cardMetrics)
	findCardByUsecase := usecase.NewFindCardByUseCase(o11y, cardRepository, cardMetrics)
	createCardUsecase := usecase.NewCreateCardUseCase(o11y, cardRepository, cardMetrics)
	updateCardUsecase := usecase.NewUpdateCardUseCase(o11y, cardRepository, cardMetrics)
	removeCardUsecase := usecase.NewRemoveCardUseCase(o11y, cardRepository, cardMetrics)

	cardHandler := http.NewCardHandler(
		o11y,
		errorHandler,
		findCardUsecase,
		createCardUsecase,
		findCardByUsecase,
		updateCardUsecase,
		removeCardUsecase,
	)

	cardRouter := http.NewCardRouter(cardHandler, authMiddleware)
	cardProvider := adapters.NewCardProviderAdapter(cardRepository, o11y)

	return CardModule{
		CardRouter:   cardRouter,
		CardProvider: cardProvider,
	}
}
