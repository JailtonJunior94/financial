package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	pmdomain "github.com/jailtonjunior94/financial/internal/payment_method/domain"

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
		span.RecordError(err)

		return nil, err
	}

	paymentMethod, err := u.repository.FindByID(ctx, paymentMethodID)
	if err != nil {
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "error finding payment method by id",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	if paymentMethod == nil {
		span.RecordError(pmdomain.ErrPaymentMethodNotFound)
		u.o11y.Logger().Error(ctx, "payment method not found",
			observability.Error(pmdomain.ErrPaymentMethodNotFound),
			observability.String("payment_method_id", id))
		return nil, pmdomain.ErrPaymentMethodNotFound
	}

	if err := paymentMethod.Update(input.Name, input.Description); err != nil {
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "error validating payment method update",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	if err := u.repository.Update(ctx, paymentMethod); err != nil {
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "error updating payment method in repository",
			observability.Error(err),
			observability.String("payment_method_id", id))
		return nil, err
	}

	return toPaymentMethodOutput(paymentMethod), nil
}
