package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/factories"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateCardUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error)
	}

	createCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewCreateCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) CreateCardUseCase {
	return &createCardUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *createCardUseCase) Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_card_usecase.execute")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			u.metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration)
			panic(err)
		}
	}()

	card, err := factories.CreateCard(userID, input.Name, input.DueDay)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration)
		u.metrics.RecordError(ctx, metrics.OperationCreate, metrics.ClassifyError(err))

		span.AddEvent(
			"error creating card entity",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	if err := u.repository.Save(ctx, card); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration)
		u.metrics.RecordError(ctx, metrics.OperationCreate, metrics.ClassifyError(err))

		span.AddEvent(
			"error saving card to repository",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationCreate, duration)
	u.metrics.IncActiveCards(ctx)

	return &dtos.CardOutput{
		ID:                card.ID.String(),
		Name:              card.Name.String(),
		DueDay:            card.DueDay.Int(),
		ClosingOffsetDays: card.ClosingOffsetDays.Int(),
		CreatedAt:         card.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
