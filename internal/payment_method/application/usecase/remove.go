package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	pmdomain "github.com/jailtonjunior94/financial/internal/payment_method/domain"
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
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, "error parsing payment method id")
		return err
	}

	paymentMethod, err := u.repository.FindByID(ctx, paymentMethodID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, "error finding payment method by id")
		u.o11y.Logger().Error(ctx, "error finding payment method by id",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return err
	}

	if paymentMethod == nil {
		span.RecordError(pmdomain.ErrPaymentMethodNotFound)
		span.SetStatus(observability.StatusCodeError, "payment method not found")
		u.o11y.Logger().Error(ctx, "payment method not found",
			observability.Error(pmdomain.ErrPaymentMethodNotFound),
			observability.String("payment_method_id", id))
		return pmdomain.ErrPaymentMethodNotFound
	}

	if err := u.repository.Update(ctx, paymentMethod.Delete()); err != nil {
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, "error deleting payment method in repository")
		u.o11y.Logger().Error(ctx, "error deleting payment method in repository",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return err
	}

	return nil
}
