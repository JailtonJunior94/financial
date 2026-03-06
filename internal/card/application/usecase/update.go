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
		Execute(ctx context.Context, userID, id string, input *dtos.CardUpdateInput) (*dtos.CardOutput, error)
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

func (u *updateCardUseCase) Execute(ctx context.Context, userID, id string, input *dtos.CardUpdateInput) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_card_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))
		span.AddEvent("error parsing user id",
			observability.String("user_id", userID),
			observability.Error(err),
		)
		return nil, err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))
		span.AddEvent("error parsing card id",
			observability.String("card_id", id),
			observability.Error(err),
		)
		return nil, err
	}

	card, err := u.repository.FindByIDOnly(ctx, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "UpdateCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		return nil, err
	}

	if card == nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ErrorTypeNotFound)
		span.AddEvent("card not found",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Warn(ctx, "card not found",
			observability.String("operation", "UpdateCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, customErrors.ErrCardNotFound
	}

	if card.UserID.String() != user.String() {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, "authorization")
		span.AddEvent("card ownership mismatch",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Warn(ctx, "card ownership mismatch",
			observability.String("operation", "UpdateCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, customErrors.ErrForbidden
	}

	dueDay := card.DueDay.Int()
	if input.DueDay != nil {
		dueDay = *input.DueDay
	}

	closingOffsetDays := card.ClosingOffsetDays.Int()
	if input.ClosingOffsetDays != nil {
		closingOffsetDays = *input.ClosingOffsetDays
	}

	if err := card.Update(input.Name, input.Flag, input.LastFourDigits, dueDay, closingOffsetDays); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "UpdateCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		return nil, err
	}

	if err := u.repository.Update(ctx, card); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationUpdate, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "UpdateCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		return nil, err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationUpdate, duration)

	output := &dtos.CardOutput{
		ID:             card.ID.String(),
		Name:           card.Name.String(),
		Type:           card.Type.Value,
		Flag:           card.Flag.Value,
		LastFourDigits: card.LastFourDigits.Value,
		CreatedAt:      card.CreatedAt.ValueOr(time.Time{}),
	}
	if card.Type.IsCredit() {
		dueDay := card.DueDay.Int()
		output.DueDay = &dueDay
		offset := card.ClosingOffsetDays.Int()
		output.ClosingOffsetDays = &offset
	}
	if !card.UpdatedAt.ValueOr(time.Time{}).IsZero() {
		output.UpdatedAt = card.UpdatedAt.ValueOr(time.Time{})
	}

	return output, nil
}
