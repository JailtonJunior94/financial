package handlers

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/events"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	invoiceevents "github.com/jailtonjunior94/financial/internal/invoice/domain/events"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// PurchaseEventHandler lida com eventos de purchase e sincroniza MonthlyTransaction.
// Implementa github.com/JailtonJunior94/devkit-go/pkg/events.EventHandler.
type PurchaseEventHandler struct {
	syncUseCase usecase.SyncMonthlyFromInvoicesUseCase
}

// NewPurchaseEventHandler cria um novo handler de eventos de purchase.
func NewPurchaseEventHandler(syncUseCase usecase.SyncMonthlyFromInvoicesUseCase) *PurchaseEventHandler {
	return &PurchaseEventHandler{
		syncUseCase: syncUseCase,
	}
}

// Handle implementa events.EventHandler interface.
func (h *PurchaseEventHandler) Handle(ctx context.Context, event events.Event) error {
	switch e := event.(type) {
	case *invoiceevents.PurchaseCreated:
		payload := e.GetPayload().(invoiceevents.PurchaseEventPayload)
		return h.syncAffectedMonths(ctx, payload.UserID, payload.CategoryID, payload.AffectedMonths)

	case *invoiceevents.PurchaseUpdated:
		payload := e.GetPayload().(invoiceevents.PurchaseEventPayload)
		return h.syncAffectedMonths(ctx, payload.UserID, payload.CategoryID, payload.AffectedMonths)

	case *invoiceevents.PurchaseDeleted:
		payload := e.GetPayload().(invoiceevents.PurchaseEventPayload)
		return h.syncAffectedMonths(ctx, payload.UserID, payload.CategoryID, payload.AffectedMonths)

	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// syncAffectedMonths sincroniza todos os meses afetados.
func (h *PurchaseEventHandler) syncAffectedMonths(ctx context.Context, userID, categoryID string, affectedMonths []string) error {
	// Converter userID string para VO
	userIDVO, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	// Converter categoryID string para VO
	categoryIDVO, err := vos.NewUUIDFromString(categoryID)
	if err != nil {
		return fmt.Errorf("invalid category_id: %w", err)
	}

	// Sincronizar cada mês afetado
	for _, month := range affectedMonths {
		referenceMonth, err := transactionVos.NewReferenceMonthFromString(month)
		if err != nil {
			return fmt.Errorf("invalid reference_month %s: %w", month, err)
		}

		if err := h.syncUseCase.Execute(ctx, userIDVO, referenceMonth, categoryIDVO); err != nil {
			return fmt.Errorf("failed to sync month %s: %w", month, err)
		}
	}

	return nil
}
