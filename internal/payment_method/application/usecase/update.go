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
	UpdatePaymentMethodUseCase interface {
		Execute(ctx context.Context, id string, input *dtos.PaymentMethodUpdateInput) (*dtos.PaymentMethodOutput, error)
	}

	updatePaymentMethodUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewUpdatePaymentMethodUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) UpdatePaymentMethodUseCase {
	return &updatePaymentMethodUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *updatePaymentMethodUseCase) Execute(ctx context.Context, id string, input *dtos.PaymentMethodUpdateInput) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_payment_method_usecase.execute")
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

	if err := paymentMethod.Update(input.Name, input.Description); err != nil {
		span.AddEvent(
			"error validating payment method update",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error validating payment method update",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	if err := u.repository.Update(ctx, paymentMethod); err != nil {
		span.AddEvent(
			"error updating payment method in repository",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error updating payment method in repository",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	output := &dtos.PaymentMethodOutput{
		ID:          paymentMethod.ID.String(),
		Name:        paymentMethod.Name.String(),
		Code:        paymentMethod.Code.String(),
		Description: paymentMethod.Description.String(),
	}
	if !paymentMethod.UpdatedAt.ValueOr(time.Time{}).IsZero() {
		output.UpdatedAt = paymentMethod.UpdatedAt.ValueOr(time.Time{})
	}

	return output, nil
}
