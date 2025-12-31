package card

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/card/application/usecase"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
)

type CardModule struct {
	CardRouter *http.CardRouter
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

	return CardModule{
		CardRouter: cardRouter,
	}
}
