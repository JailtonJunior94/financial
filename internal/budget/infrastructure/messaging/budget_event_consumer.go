package messaging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/google/uuid"

	invoiceevents "github.com/jailtonjunior94/financial/internal/invoice/domain/events"

	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

// purchaseEventPayload espelha o payload publicado pelo módulo de invoice.
type purchaseEventPayload struct {
	Version        int      `json:"version"`
	UserID         string   `json:"user_id"`
	CategoryID     string   `json:"category_id"`
	AffectedMonths []string `json:"affected_months"`
}

// BudgetEventConsumer processa eventos de purchase e sincroniza o spent_amount do budget.
type BudgetEventConsumer struct {
	syncUseCase         usecase.SyncBudgetSpentAmountUseCase
	db                  *sql.DB
	processedEventsRepo outbox.ProcessedEventsRepository
	o11y                observability.Observability
}

func NewBudgetEventConsumer(
	syncUseCase usecase.SyncBudgetSpentAmountUseCase,
	db *sql.DB,
	o11y observability.Observability,
) *BudgetEventConsumer {
	return &BudgetEventConsumer{
		syncUseCase:         syncUseCase,
		db:                  db,
		processedEventsRepo: outbox.NewProcessedEventsRepository(db),
		o11y:                o11y,
	}
}

// Handle implementa messaging.Handler.
func (c *BudgetEventConsumer) Handle(ctx context.Context, msg *messaging.Message) error {
	ctx, span := c.o11y.Tracer().Start(ctx, "budget_event_consumer.handle")
	defer span.End()

	const consumerName = "budget_event_consumer"

	eventID, err := uuid.Parse(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid message ID format: %w", err)
	}

	processed, err := c.processedEventsRepo.IsProcessed(ctx, eventID, consumerName)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}

	if processed {
		c.o11y.Logger().Info(ctx, "budget event already processed, skipping",
			observability.String("event_id", eventID.String()),
		)
		return nil
	}

	var payload purchaseEventPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal budget event payload: %w", err)
	}

	userID, err := vos.NewUUIDFromString(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	categoryID, err := vos.NewUUIDFromString(payload.CategoryID)
	if err != nil {
		return fmt.Errorf("invalid category_id: %w", err)
	}

	var syncErrors []error

	for _, month := range payload.AffectedMonths {
		referenceMonth, err := pkgVos.NewReferenceMonth(month)
		if err != nil {
			c.o11y.Logger().Error(ctx, "invalid reference month in budget event",
				observability.Error(err),
				observability.String("month", month),
			)
			syncErrors = append(syncErrors, fmt.Errorf("invalid month %s: %w", month, err))
			continue
		}

		if err := c.syncUseCase.Execute(ctx, userID, referenceMonth, categoryID); err != nil {
			c.o11y.Logger().Error(ctx, "failed to sync budget spent amount",
				observability.Error(err),
				observability.String("month", month),
			)
			syncErrors = append(syncErrors, fmt.Errorf("failed to sync month %s: %w", month, err))
			continue
		}

		c.o11y.Logger().Info(ctx, "budget synced for month",
			observability.String("user_id", payload.UserID),
			observability.String("month", month),
			observability.String("category_id", payload.CategoryID),
		)
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("failed to sync %d of %d months: %v", len(syncErrors), len(payload.AffectedMonths), syncErrors)
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for marking event: %w", err)
	}
	defer tx.Rollback()

	processedRepo := outbox.NewProcessedEventsRepository(tx)
	if err := processedRepo.MarkAsProcessed(ctx, eventID, consumerName); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit processed event: %w", err)
	}

	c.o11y.Logger().Info(ctx, "budget event processed successfully",
		observability.String("event_id", eventID.String()),
		observability.Int("months_synced", len(payload.AffectedMonths)),
	)

	return nil
}

// Topics retorna as routing keys que este consumer processa.
func (c *BudgetEventConsumer) Topics() []string {
	return []string{
		"invoice." + invoiceevents.PurchaseCreatedEventName,
		"invoice." + invoiceevents.PurchaseUpdatedEventName,
		"invoice." + invoiceevents.PurchaseDeletedEventName,
	}
}
