package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindPaymentMethodByUseCase interface {
		Execute(ctx context.Context, id string) (*dtos.PaymentMethodOutput, error)
	}

	findPaymentMethodByUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewFindPaymentMethodByUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) FindPaymentMethodByUseCase {
	return &findPaymentMethodByUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findPaymentMethodByUseCase) Execute(ctx context.Context, id string) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_payment_method_by_usecase.execute")
	defer span.End()

	paymentMethodID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing payment method id",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	paymentMethod, err := u.repository.FindByID(ctx, paymentMethodID)
	if err != nil {
		span.AddEvent(
			"error finding payment method by id",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error finding payment method by id",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	if paymentMethod == nil {
		span.AddEvent(
			"payment method not found",
			observability.Field{Key: "payment_method_id", Value: id},
		)
		u.o11y.Logger().Error(ctx, "payment method not found",
			observability.Error(customErrors.ErrPaymentMethodNotFound),
			observability.String("payment_method_id", id))
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
