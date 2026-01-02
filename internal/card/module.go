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

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type CardModule struct {
	CardRouter   *http.CardRouter
	CardProvider invoiceInterfaces.CardProvider // ✅ Export adapter for invoice module
}

func NewCardModule(db *sql.DB, o11y observability.Observability, tokenValidator auth.TokenValidator) CardModule {
	errorHandler := httperrors.NewErrorHandler(o11y)
	authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)

	cardRepository := repositories.NewCardRepository(db, o11y)
	findCardUsecase := usecase.NewFindCardUseCase(o11y, cardRepository)
	findCardByUsecase := usecase.NewFindCardByUseCase(o11y, cardRepository)
	createCardUsecase := usecase.NewCreateCardUseCase(o11y, cardRepository)
	updateCardUsecase := usecase.NewUpdateCardUseCase(o11y, cardRepository)
	removeCardUsecase := usecase.NewRemoveCardUseCase(o11y, cardRepository)

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

	// ✅ Create CardProvider adapter for invoice module
	cardProvider := adapters.NewCardProviderAdapter(cardRepository, o11y)

	return CardModule{
		CardRouter:   cardRouter,
		CardProvider: cardProvider,
	}
}
