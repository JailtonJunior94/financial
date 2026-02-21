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
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

// PurchaseEventConsumer processa eventos de purchase vindos do RabbitMQ.
// Implementa messaging.Handler.
type PurchaseEventConsumer struct {
	syncUseCase         usecase.SyncMonthlyFromInvoicesUseCase
	processedEventsRepo outbox.ProcessedEventsRepository
	o11y                observability.Observability
}

// NewPurchaseEventConsumer cria um novo consumer de eventos de purchase.
func NewPurchaseEventConsumer(
	syncUseCase usecase.SyncMonthlyFromInvoicesUseCase,
	db *sql.DB,
	o11y observability.Observability,
) *PurchaseEventConsumer {
	return &PurchaseEventConsumer{
		syncUseCase:         syncUseCase,
		processedEventsRepo: outbox.NewProcessedEventsRepository(db),
		o11y:                o11y,
	}
}

// Handle implementa messaging.Handler interface.
func (c *PurchaseEventConsumer) Handle(ctx context.Context, msg *messaging.Message) error {
	ctx, span := c.o11y.Tracer().Start(ctx, "purchase_event_consumer.handle")
	defer span.End()

	const consumerName = "purchase_event_consumer"

	// Parse message ID como UUID (event_id do outbox)
	eventID, err := uuid.Parse(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid message ID format: %w", err)
	}

	// ✅ IDEMPOTÊNCIA ATÔMICA: TryClaimEvent usa INSERT ... ON CONFLICT DO NOTHING,
	// garantindo que apenas um worker processa o evento mesmo com consumers concorrentes.
	// Elimina o race condition TOCTOU da abordagem anterior (IsProcessed + MarkAsProcessed).
	claimed, err := c.processedEventsRepo.TryClaimEvent(ctx, eventID, consumerName)
	if err != nil {
		c.o11y.Logger().Error(ctx, "failed to claim event",
			observability.Error(err),
			observability.String("event_id", eventID.String()),
		)
		return fmt.Errorf("failed to claim event: %w", err)
	}

	if !claimed {
		c.o11y.Logger().Info(ctx, "event already processed, skipping",
			observability.String("event_id", eventID.String()),
		)
		return nil
	}

	// Extrair event_type do header
	eventType, ok := msg.Headers["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type header")
	}

	c.o11y.Logger().Info(ctx, "processing purchase event",
		observability.String("event_type", eventType),
		observability.String("topic", msg.Topic),
		observability.String("event_id", eventID.String()),
	)

	// Parse payload usando o contrato canônico definido no módulo de invoice
	var payload invoiceevents.PurchaseEventPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal purchase event: %w", err)
	}

	// Converter userID para VO
	userIDVO, err := vos.NewUUIDFromString(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	// Converter categoryID para VO
	categoryIDVO, err := vos.NewUUIDFromString(payload.CategoryID)
	if err != nil {
		return fmt.Errorf("invalid category_id: %w", err)
	}

	// ✅ CORREÇÃO: Coletar erros de todos os meses e retornar se algum falhar
	var syncErrors []error

	// Sincronizar cada mês afetado
	for _, month := range payload.AffectedMonths {
		referenceMonth, err := pkgVos.NewReferenceMonth(month)
		if err != nil {
			c.o11y.Logger().Error(ctx, "invalid reference month",
				observability.Error(err),
				observability.String("month", month),
			)
			syncErrors = append(syncErrors, fmt.Errorf("invalid month %s: %w", month, err))
			continue
		}

		if err := c.syncUseCase.Execute(ctx, userIDVO, referenceMonth, categoryIDVO); err != nil {
			c.o11y.Logger().Error(ctx, "failed to sync monthly transaction",
				observability.Error(err),
				observability.String("month", month),
			)
			syncErrors = append(syncErrors, fmt.Errorf("failed to sync month %s: %w", month, err))
			continue
		}

		c.o11y.Logger().Info(ctx, "monthly transaction synced successfully",
			observability.String("user_id", payload.UserID),
			observability.String("month", month),
			observability.String("event_type", eventType),
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
	c.o11y.Logger().Info(ctx, "purchase event processed successfully",
		observability.String("event_id", eventID.String()),
		observability.Int("months_synced", len(payload.AffectedMonths)),
	)

	return nil
}

// Topics implementa messaging.Handler interface.
// Retorna routing keys que este handler processa.
// As routing keys já incluem o aggregate_type prefix (invoice.purchase.*)
func (c *PurchaseEventConsumer) Topics() []string {
	return []string{
		"invoice." + invoiceevents.PurchaseCreatedEventName, // invoice.purchase.created
		"invoice." + invoiceevents.PurchaseUpdatedEventName, // invoice.purchase.updated
		"invoice." + invoiceevents.PurchaseDeletedEventName, // invoice.purchase.deleted
	}
}
