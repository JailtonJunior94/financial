package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/google/uuid"

	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

// BudgetEventConsumer consumes transaction.created events and syncs budget spent amounts.
type BudgetEventConsumer struct {
	syncUseCase         usecase.SyncBudgetSpentAmountUseCase
	processedEventsRepo outbox.ProcessedEventsRepository
	o11y                observability.Observability
}

// NewBudgetEventConsumer creates a BudgetEventConsumer with injected dependencies.
func NewBudgetEventConsumer(
	syncUseCase usecase.SyncBudgetSpentAmountUseCase,
	processedEventsRepo outbox.ProcessedEventsRepository,
	o11y observability.Observability,
) *BudgetEventConsumer {
	return &BudgetEventConsumer{
		syncUseCase:         syncUseCase,
		processedEventsRepo: processedEventsRepo,
		o11y:                o11y,
	}
}

// transactionCreatedPayload mirrors the TransactionCreatedEvent payload contract.
type transactionCreatedPayload struct {
	TransactionID  string `json:"transaction_id"`
	UserID         string `json:"user_id"`
	CategoryID     string `json:"category_id"`
	ReferenceMonth string `json:"reference_month"`
}

// Handle implements messaging.Handler for transaction.created events.
func (c *BudgetEventConsumer) Handle(ctx context.Context, msg *messaging.Message) error {
	ctx, span := c.o11y.Tracer().Start(ctx, "budget_event_consumer.handle")
	defer span.End()

	const consumerName = "budget_event_consumer"

	eventID, err := uuid.Parse(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid message ID format: %w", err)
	}

	claimed, err := c.processedEventsRepo.TryClaimEvent(ctx, eventID, consumerName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to claim event: %w", err)
	}

	if !claimed {
		c.o11y.Logger().Info(ctx, "event_already_processed",
			observability.String("operation", "handle_transaction_created"),
			observability.String("layer", "consumer"),
			observability.String("entity", "budget"),
			observability.String("event_id", eventID.String()),
		)
		return nil
	}

	var payload transactionCreatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to parse transaction.created payload: %w", err)
	}

	userID, err := vos.NewUUIDFromString(payload.UserID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid user_id: %w", err)
	}

	categoryID, err := vos.NewUUIDFromString(payload.CategoryID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid category_id: %w", err)
	}

	referenceMonth, err := pkgVos.NewReferenceMonth(payload.ReferenceMonth)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid reference_month: %w", err)
	}

	if err := c.syncUseCase.Execute(ctx, userID, referenceMonth, categoryID); err != nil {
		span.RecordError(err)
		if deleteErr := c.processedEventsRepo.DeleteClaim(ctx, eventID, consumerName); deleteErr != nil {
			c.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "delete_claim"),
				observability.String("layer", "consumer"),
				observability.String("entity", "budget"),
				observability.String("event_id", eventID.String()),
				observability.Error(deleteErr),
			)
		}
		return fmt.Errorf("failed to sync budget: %w", err)
	}

	c.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "handle_transaction_created"),
		observability.String("layer", "consumer"),
		observability.String("entity", "budget"),
		observability.String("event_id", eventID.String()),
		observability.String("user_id", payload.UserID),
		observability.String("reference_month", payload.ReferenceMonth),
	)

	return nil
}

// Topics returns the routing keys this consumer handles.
func (c *BudgetEventConsumer) Topics() []string {
	return []string{"transaction.created"}
}
