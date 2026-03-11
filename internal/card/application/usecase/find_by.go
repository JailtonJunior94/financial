package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	cardDomain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCardByUseCase interface {
		Execute(ctx context.Context, userID, id string) (*dtos.CardOutput, error)
	}

	findCardByUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewFindCardByUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) FindCardByUseCase {
	return &findCardByUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *findCardByUseCase) Execute(ctx context.Context, userID, id string) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_card_by_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		return nil, err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		return nil, err
	}

	card, err := u.repository.FindByIDOnly(ctx, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "FindCardBy"),
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
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ErrorTypeNotFound)
		span.RecordError(cardDomain.ErrCardNotFound)
		u.o11y.Logger().Warn(ctx, "card not found",
			observability.String("operation", "FindCardBy"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, cardDomain.ErrCardNotFound
	}

	if card.UserID.String() != user.String() {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, "authorization")
		span.RecordError(customErrors.ErrForbidden)
		u.o11y.Logger().Warn(ctx, "card ownership mismatch",
			observability.String("operation", "FindCardBy"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, customErrors.ErrForbidden
	}

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

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationFindBy, duration)
	return output, nil
}
