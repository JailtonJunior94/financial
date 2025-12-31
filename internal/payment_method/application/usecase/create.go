package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/factories"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreatePaymentMethodUseCase interface {
		Execute(ctx context.Context, input *dtos.PaymentMethodInput) (*dtos.PaymentMethodOutput, error)
	}

	createPaymentMethodUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewCreatePaymentMethodUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) CreatePaymentMethodUseCase {
	return &createPaymentMethodUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *createPaymentMethodUseCase) Execute(ctx context.Context, input *dtos.PaymentMethodInput) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_payment_method_usecase.execute")
	defer span.End()

	paymentMethod, err := factories.CreatePaymentMethod(input.Name, input.Code, input.Description)
	if err != nil {
		span.AddEvent(
			"error creating payment method entity",
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	if err := u.repository.Save(ctx, paymentMethod); err != nil {
		span.AddEvent(
			"error saving payment method to repository",
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	output := &dtos.PaymentMethodOutput{
		ID:          paymentMethod.ID.String(),
		Name:        paymentMethod.Name.String(),
		Code:        paymentMethod.Code.String(),
		Description: paymentMethod.Description.String(),
		CreatedAt:   paymentMethod.CreatedAt.ValueOr(time.Time{}),
	}

	return output, nil
}
