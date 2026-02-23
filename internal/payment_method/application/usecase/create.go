package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/factories"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"go.opentelemetry.io/otel/trace"
)

type (
	CreatePaymentMethodUseCase interface {
		Execute(ctx context.Context, input *dtos.PaymentMethodInput) (*dtos.PaymentMethodOutput, error)
	}

	createPaymentMethodUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
		fm         *metrics.FinancialMetrics
	}
)

func NewCreatePaymentMethodUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
	fm *metrics.FinancialMetrics,
) CreatePaymentMethodUseCase {
	return &createPaymentMethodUseCase{
		o11y:       o11y,
		repository: repository,
		fm:         fm,
	}
}

func (u *createPaymentMethodUseCase) Execute(ctx context.Context, input *dtos.PaymentMethodInput) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_payment_method_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "execution_started",
		observability.String("operation", "CreatePaymentMethod"),
		observability.String("layer", "usecase"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
	)

	paymentMethod, err := factories.CreatePaymentMethod(input.Name, input.Code, input.Description)
	if err != nil {
		u.fm.RecordUsecaseFailure(ctx, "CreatePaymentMethod", "payment_method", "validation", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "CreatePaymentMethod"),
			observability.String("layer", "usecase"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "CREATE_PAYMENT_METHOD_ENTITY_FAILED"),
			observability.Error(err),
		)
		span.AddEvent(
			"error creating payment method entity",
			observability.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	if err := u.repository.Save(ctx, paymentMethod); err != nil {
		u.fm.RecordUsecaseFailure(ctx, "CreatePaymentMethod", "payment_method", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "CreatePaymentMethod"),
			observability.String("layer", "usecase"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "SAVE_PAYMENT_METHOD_FAILED"),
			observability.Error(err),
		)
		span.AddEvent(
			"error saving payment method to repository",
			observability.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	u.fm.RecordUsecaseOperation(ctx, "CreatePaymentMethod", "payment_method", time.Since(start))
	u.o11y.Logger().Info(ctx, "execution_completed",
		observability.String("operation", "CreatePaymentMethod"),
		observability.String("layer", "usecase"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", paymentMethod.ID.String()),
	)

	output := &dtos.PaymentMethodOutput{
		ID:          paymentMethod.ID.String(),
		Name:        paymentMethod.Name.String(),
		Code:        paymentMethod.Code.String(),
		Description: paymentMethod.Description.String(),
		CreatedAt:   paymentMethod.CreatedAt.ValueOr(time.Time{}),
	}

	return output, nil
}
