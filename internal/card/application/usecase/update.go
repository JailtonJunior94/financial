package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateCardUseCase interface {
		Execute(ctx context.Context, userID, id string, input *dtos.CardInput) (*dtos.CardOutput, error)
	}

	updateCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewUpdateCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) UpdateCardUseCase {
	return &updateCardUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *updateCardUseCase) Execute(ctx context.Context, userID, id string, input *dtos.CardInput) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_card_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing user id",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing card id",
			observability.String("card_id", id),
			observability.Error(err),
		)

		return nil, err
	}

	card, err := u.repository.FindByID(ctx, user, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error finding card by id",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(
			ctx,
			"error finding card by id",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, err
	}

	if card == nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"card not found",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Error(
			ctx,
			"card not found",
			observability.Error(customErrors.ErrCardNotFound),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, customErrors.ErrCardNotFound
	}

	// Se não fornecido, mantém o valor atual
	closingOffsetDays := input.ClosingOffsetDays
	if closingOffsetDays == 0 {
		closingOffsetDays = card.ClosingOffsetDays.Int()
	}

	if err := card.Update(input.Name, input.DueDay, closingOffsetDays); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error validating card update",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(
			ctx,
			"error validating card update",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, err
	}

	if err := u.repository.Update(ctx, card); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error updating card in repository",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(
			ctx,
			"error updating card in repository",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationUpdate, duration)

	output := &dtos.CardOutput{
		ID:                card.ID.String(),
		Name:              card.Name.String(),
		DueDay:            card.DueDay.Int(),
		ClosingOffsetDays: card.ClosingOffsetDays.Int(),
	}
	if !card.UpdatedAt.ValueOr(time.Time{}).IsZero() {
		output.UpdatedAt = card.UpdatedAt.ValueOr(time.Time{})
	}

	return output, nil
}
