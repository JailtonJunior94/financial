package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

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
			observability.Error(customErrors.ErrPaymentMethodNotFound),
			observability.String("code", code))
		return nil, customErrors.ErrPaymentMethodNotFound
	}

	output := &dtos.PaymentMethodOutput{
		ID:          paymentMethod.ID.String(),
		Name:        paymentMethod.Name.String(),
		Code:        paymentMethod.Code.String(),
		Description: paymentMethod.Description.String(),
		CreatedAt:   paymentMethod.CreatedAt.ValueOr(time.Time{}),
	}
	if !paymentMethod.UpdatedAt.ValueOr(time.Time{}).IsZero() {
		output.UpdatedAt = paymentMethod.UpdatedAt.ValueOr(time.Time{})
	}

	return output, nil
}
