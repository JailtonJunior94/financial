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

// PurchaseEventPayload representa o payload dos eventos de purchase do outbox.
type PurchaseEventPayload struct {
	Version        int      `json:"version"`
	UserID         string   `json:"user_id"`
	CategoryID     string   `json:"category_id"`
	AffectedMonths []string `json:"affected_months"`
}

// PurchaseEventConsumer processa eventos de purchase vindos do RabbitMQ.
// Implementa messaging.Handler.
type PurchaseEventConsumer struct {
	syncUseCase       usecase.SyncMonthlyFromInvoicesUseCase
	db                *sql.DB
	processedEventsRepo outbox.ProcessedEventsRepository
	o11y              observability.Observability
}

// NewPurchaseEventConsumer cria um novo consumer de eventos de purchase.
func NewPurchaseEventConsumer(
	syncUseCase usecase.SyncMonthlyFromInvoicesUseCase,
	db *sql.DB,
	o11y observability.Observability,
) *PurchaseEventConsumer {
	return &PurchaseEventConsumer{
		syncUseCase:         syncUseCase,
		db:                  db,
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

	// ✅ IDEMPOTÊNCIA: Verificar se evento já foi processado
	processed, err := c.processedEventsRepo.IsProcessed(ctx, eventID, consumerName)
	if err != nil {
		c.o11y.Logger().Error(ctx, "failed to check if event is processed",
			observability.Error(err),
			observability.String("event_id", eventID.String()),
		)
		return fmt.Errorf("failed to check idempotency: %w", err)
	}

	if processed {
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

	// Parse payload
	var payload PurchaseEventPayload
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

	// ✅ CORREÇÃO: Se houve erros, retornar para forçar retry da mensagem
	if len(syncErrors) > 0 {
		return fmt.Errorf("failed to sync %d of %d months: %v", len(syncErrors), len(payload.AffectedMonths), syncErrors)
	}

	// ✅ IDEMPOTÊNCIA: Marcar evento como processado após sucesso
	// Usar transação separada para garantir persistência mesmo se o processo morrer
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		c.o11y.Logger().Error(ctx, "failed to begin transaction for marking event",
			observability.Error(err),
		)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	processedRepo := outbox.NewProcessedEventsRepository(tx)
	if err := processedRepo.MarkAsProcessed(ctx, eventID, consumerName); err != nil {
		c.o11y.Logger().Error(ctx, "failed to mark event as processed",
			observability.Error(err),
			observability.String("event_id", eventID.String()),
		)
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		c.o11y.Logger().Error(ctx, "failed to commit processed event",
			observability.Error(err),
		)
		return fmt.Errorf("failed to commit: %w", err)
	}

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
