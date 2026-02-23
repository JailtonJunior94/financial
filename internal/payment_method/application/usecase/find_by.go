package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"go.opentelemetry.io/otel/trace"
)

type (
	FindPaymentMethodByUseCase interface {
		Execute(ctx context.Context, id string) (*dtos.PaymentMethodOutput, error)
	}

	findPaymentMethodByUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
		fm         *metrics.FinancialMetrics
	}
)

func NewFindPaymentMethodByUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
	fm *metrics.FinancialMetrics,
) FindPaymentMethodByUseCase {
	return &findPaymentMethodByUseCase{
		o11y:       o11y,
		repository: repository,
		fm:         fm,
	}
}

func (u *findPaymentMethodByUseCase) Execute(ctx context.Context, id string) (*dtos.PaymentMethodOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_payment_method_by_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "execution_started",
		observability.String("operation", "FindPaymentMethodBy"),
		observability.String("layer", "usecase"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
	)

	paymentMethodID, err := vos.NewUUIDFromString(id)
	if err != nil {
		u.fm.RecordUsecaseFailure(ctx, "FindPaymentMethodBy", "payment_method", "validation", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "FindPaymentMethodBy"),
			observability.String("layer", "usecase"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "INVALID_PAYMENT_METHOD_ID"),
			observability.String("payment_method_id", id),
			observability.Error(err),
		)
		span.AddEvent(
			"error parsing payment method id",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	paymentMethod, err := u.repository.FindByID(ctx, paymentMethodID)
	if err != nil {
		u.fm.RecordUsecaseFailure(ctx, "FindPaymentMethodBy", "payment_method", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "FindPaymentMethodBy"),
			observability.String("layer", "usecase"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "FIND_PAYMENT_METHOD_FAILED"),
			observability.String("payment_method_id", id),
			observability.Error(err),
		)
		span.AddEvent(
			"error finding payment method by id",
			observability.Field{Key: "payment_method_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	if paymentMethod == nil {
		u.fm.RecordUsecaseFailure(ctx, "FindPaymentMethodBy", "payment_method", "business", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "FindPaymentMethodBy"),
			observability.String("layer", "usecase"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "PAYMENT_METHOD_NOT_FOUND"),
			observability.String("payment_method_id", id),
			observability.Error(customErrors.ErrPaymentMethodNotFound),
		)
		span.AddEvent(
			"payment method not found",
			observability.Field{Key: "payment_method_id", Value: id},
		)
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

	u.fm.RecordUsecaseOperation(ctx, "FindPaymentMethodBy", "payment_method", time.Since(start))
	u.o11y.Logger().Info(ctx, "execution_completed",
		observability.String("operation", "FindPaymentMethodBy"),
		observability.String("layer", "usecase"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", id),
	)

	return output, nil
}
