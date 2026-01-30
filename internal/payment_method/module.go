package payment_method

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/usecase"
	"github.com/jailtonjunior94/financial/internal/payment_method/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/payment_method/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type PaymentMethodModule struct {
	PaymentMethodRouter *http.PaymentMethodRouter
}

func NewPaymentMethodModule(db *sql.DB, o11y observability.Observability) PaymentMethodModule {
	errorHandler := httperrors.NewErrorHandler(o11y)

	paymentMethodRepository := repositories.NewPaymentMethodRepository(db, o11y)
	findPaymentMethodUsecase := usecase.NewFindPaymentMethodUseCase(o11y, paymentMethodRepository)
	findPaymentMethodPaginatedUsecase := usecase.NewFindPaymentMethodPaginatedUseCase(o11y, paymentMethodRepository)
	findPaymentMethodByUsecase := usecase.NewFindPaymentMethodByUseCase(o11y, paymentMethodRepository)
	findPaymentMethodByCodeUsecase := usecase.NewFindPaymentMethodByCodeUseCase(o11y, paymentMethodRepository)
	createPaymentMethodUsecase := usecase.NewCreatePaymentMethodUseCase(o11y, paymentMethodRepository)
	updatePaymentMethodUsecase := usecase.NewUpdatePaymentMethodUseCase(o11y, paymentMethodRepository)
	removePaymentMethodUsecase := usecase.NewRemovePaymentMethodUseCase(o11y, paymentMethodRepository)

	paymentMethodHandler := http.NewPaymentMethodHandler(
		o11y,
		errorHandler,
		findPaymentMethodUsecase,
		findPaymentMethodPaginatedUsecase,
		createPaymentMethodUsecase,
		findPaymentMethodByUsecase,
		findPaymentMethodByCodeUsecase,
		updatePaymentMethodUsecase,
		removePaymentMethodUsecase,
	)

	paymentMethodRouter := http.NewPaymentMethodRouter(paymentMethodHandler)

	return PaymentMethodModule{
		PaymentMethodRouter: paymentMethodRouter,
	}
}
