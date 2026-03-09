package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	pmdomain "github.com/jailtonjunior94/financial/internal/payment_method/domain"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	FindPaymentMethodByCodeUseCase interface {
		Execute(ctx context.Context, code string) (*dtos.PaymentMethodOutput, error)
	}

	findPaymentMethodByCodeUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewFindPaymentMethodByCodeUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) FindPaymentMethodByCodeUseCase {
	return &findPaymentMethodByCodeUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findPaymentMethodByCodeUseCase) Execute(ctx context.Context, code string) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_payment_method_by_code_usecase.execute")
	defer span.End()

	paymentMethod, err := u.repository.FindByCode(ctx, code)
	if err != nil {
		span.AddEvent(
			"error finding payment method by code",
			observability.Field{Key: "code", Value: code},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error finding payment method by code",
			observability.Error(err),
			observability.String("code", code))
		return nil, err
	}

	if paymentMethod == nil {
		span.AddEvent(
			"payment method not found",
			observability.Field{Key: "code", Value: code},
		)
		u.o11y.Logger().Error(ctx, "payment method not found",
			observability.Error(pmdomain.ErrPaymentMethodNotFound),
			observability.String("code", code))
		return nil, pmdomain.ErrPaymentMethodNotFound
	}

	return toPaymentMethodOutput(paymentMethod), nil
}
