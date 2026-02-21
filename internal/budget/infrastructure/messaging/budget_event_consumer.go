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

// BudgetEventConsumer processa eventos de purchase e sincroniza o spent_amount do budget.
type BudgetEventConsumer struct {
	syncUseCase         usecase.SyncBudgetSpentAmountUseCase
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

	// ✅ IDEMPOTÊNCIA ATÔMICA: TryClaimEvent usa INSERT ... ON CONFLICT DO NOTHING,
	// garantindo que apenas um worker processa o evento mesmo com consumers concorrentes.
	// Elimina o race condition TOCTOU da abordagem anterior (IsProcessed + MarkAsProcessed).
	claimed, err := c.processedEventsRepo.TryClaimEvent(ctx, eventID, consumerName)
	if err != nil {
		return fmt.Errorf("failed to claim event: %w", err)
	}

	if !claimed {
		c.o11y.Logger().Info(ctx, "budget event already processed, skipping",
			observability.String("event_id", eventID.String()),
		)
		return nil
	}

	// Parse payload usando o contrato canônico definido no módulo de invoice
	var payload invoiceevents.PurchaseEventPayload
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

	// Se houve erros de sincronização, liberar o claim para que a mensagem seja
	// reentregue pelo broker e processada novamente.
	if len(syncErrors) > 0 {
		if deleteErr := c.processedEventsRepo.DeleteClaim(ctx, eventID, consumerName); deleteErr != nil {
			c.o11y.Logger().Error(ctx, "failed to delete claim after sync error",
				observability.Error(deleteErr),
				observability.String("event_id", eventID.String()),
			)
		}
		return fmt.Errorf("failed to sync %d of %d months: %v", len(syncErrors), len(payload.AffectedMonths), syncErrors)
	}

	// Claim inserido pelo TryClaimEvent permanece como registro permanente de idempotência.
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
