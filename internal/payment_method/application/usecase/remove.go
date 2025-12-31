package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type (
	RemovePaymentMethodUseCase interface {
		Execute(ctx context.Context, id string) error
	}

	removePaymentMethodUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewRemovePaymentMethodUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) RemovePaymentMethodUseCase {
	return &removePaymentMethodUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *removePaymentMethodUseCase) Execute(ctx context.Context, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_payment_method_usecase.execute")
	defer span.End()

	paymentMethodID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing payment method id",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)

		return err
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
		return err
	}

	if paymentMethod == nil {
		span.AddEvent(
			"payment method not found",
			observability.Field{Key: "payment_method_id", Value: id},
		)
		u.o11y.Logger().Error(ctx, "payment method not found",
			observability.Error(customErrors.ErrPaymentMethodNotFound),
			observability.String("payment_method_id", id))
		return customErrors.ErrPaymentMethodNotFound
	}

	if err := u.repository.Update(ctx, paymentMethod.Delete()); err != nil {
		span.AddEvent(
			"error deleting payment method in repository",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error deleting payment method in repository",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return err
	}

	return nil
}
