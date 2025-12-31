package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	FindPaymentMethodUseCase interface {
		Execute(ctx context.Context) ([]*dtos.PaymentMethodOutput, error)
	}

	findPaymentMethodUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

func NewFindPaymentMethodUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) FindPaymentMethodUseCase {
	return &findPaymentMethodUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findPaymentMethodUseCase) Execute(ctx context.Context) ([]*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_payment_method_usecase.execute")
	defer span.End()

	paymentMethods, err := u.repository.List(ctx)
	if err != nil {
		span.AddEvent(
			"error listing payment methods from repository",
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	paymentMethodsOutput := linq.Map(paymentMethods, func(pm *entities.PaymentMethod) *dtos.PaymentMethodOutput {
		output := &dtos.PaymentMethodOutput{
			ID:          pm.ID.String(),
			Name:        pm.Name.String(),
			Code:        pm.Code.String(),
			Description: pm.Description.String(),
			CreatedAt:   pm.CreatedAt.ValueOr(time.Time{}),
		}
		if !pm.UpdatedAt.ValueOr(time.Time{}).IsZero() {
			output.UpdatedAt = pm.UpdatedAt.ValueOr(time.Time{})
		}
		return output
	})

	return paymentMethodsOutput, nil
}
