package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/consumer"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
)

// BudgetUpdateHandler processa eventos de transação e fatura para atualizar budgets.
//
// Este handler consome eventos:
// - transaction.transaction_created: Atualiza budget quando uma transação é criada
// - invoice.invoice_item_added: Atualiza budget quando um item de fatura é adicionado
//
// Lógica:
// 1. Extrai payload do evento
// 2. Ignora eventos de income (direction=in)
// 3. Para expense (direction=out) e invoice items, incrementa spent_amount
// 4. Busca budget do usuário/mês
// 5. Incrementa spent_amount da categoria correspondente
// 6. Recalcula percentage_used automaticamente
type BudgetUpdateHandler struct {
	incrementSpentUseCase usecase.IncrementSpentAmountUseCase
	o11y                  observability.Observability
}

// NewBudgetUpdateHandler cria um novo handler de atualização de budget.
func NewBudgetUpdateHandler(
	incrementSpentUseCase usecase.IncrementSpentAmountUseCase,
	o11y observability.Observability,
) consumer.MessageHandlerFunc {
	handler := &BudgetUpdateHandler{
		incrementSpentUseCase: incrementSpentUseCase,
		o11y:                  o11y,
	}
	return handler.Handle
}

// Handle implementa consumer.MessageHandlerFunc.
func (h *BudgetUpdateHandler) Handle(ctx context.Context, msg *consumer.Message) error {
	ctx, span := h.o11y.Tracer().Start(ctx, "budget_update_handler.handle")
	defer span.End()

	// Extract routing key from topic (devkit-go uses Topic field)
	routingKey := msg.Topic
	h.o11y.Logger().Debug(ctx, "processing budget update event",
		observability.String("topic", routingKey),
	)

	// Route to specific handler based on topic
	switch routingKey {
	case "transaction.transaction_created":
		return h.handleTransactionCreated(ctx, msg.Value)
	case "invoice.invoice_item_added":
		return h.handleInvoiceItemAdded(ctx, msg.Value)
	default:
		h.o11y.Logger().Warn(ctx, "unknown event type - ignoring",
			observability.String("topic", routingKey),
		)
		return nil // ACK unknown events
	}
}

// handleTransactionCreated processa evento de transação criada.
func (h *BudgetUpdateHandler) handleTransactionCreated(ctx context.Context, payload []byte) error {
	ctx, span := h.o11y.Tracer().Start(ctx, "budget_update_handler.handle_transaction_created")
	defer span.End()

	// Parse payload
	var event struct {
		UserID         string `json:"user_id"`
		CategoryID     string `json:"category_id"`
		Amount         int64  `json:"amount"` // in cents
		Currency       string `json:"currency"`
		Direction      string `json:"direction"`       // "in" or "out"
		ReferenceMonth string `json:"reference_month"` // "2024-01"
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		h.o11y.Logger().Error(ctx, "failed to unmarshal transaction event",
			observability.Error(err),
		)
		return fmt.Errorf("failed to unmarshal transaction event: %w", err)
	}

	// Ignore income transactions (only expenses affect budget)
	if event.Direction == "in" {
		h.o11y.Logger().Debug(ctx, "ignoring income transaction",
			observability.String("user_id", event.UserID),
		)
		return nil // ACK without processing
	}

	// Parse UUIDs
	userID, err := vos.NewUUIDFromString(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	categoryID, err := vos.NewUUIDFromString(event.CategoryID)
	if err != nil {
		return fmt.Errorf("invalid category_id: %w", err)
	}

	// Parse reference month
	referenceMonth, err := budgetVos.NewReferenceMonth(event.ReferenceMonth)
	if err != nil {
		return fmt.Errorf("invalid reference_month: %w", err)
	}

	// Parse currency
	currency, err := vos.NewCurrency(event.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	// Create Money from cents
	amount, err := vos.NewMoney(event.Amount, currency)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Increment spent amount
	if err := h.incrementSpentUseCase.Execute(ctx, userID, referenceMonth, categoryID, amount); err != nil {
		h.o11y.Logger().Error(ctx, "failed to increment spent amount",
			observability.Error(err),
			observability.String("user_id", event.UserID),
			observability.String("category_id", event.CategoryID),
		)
		return err
	}

	h.o11y.Logger().Info(ctx, "transaction event processed",
		observability.String("user_id", event.UserID),
		observability.String("category_id", event.CategoryID),
		observability.Int64("amount_cents", event.Amount),
	)

	return nil
}

// handleInvoiceItemAdded processa evento de item de fatura adicionado.
func (h *BudgetUpdateHandler) handleInvoiceItemAdded(ctx context.Context, payload []byte) error {
	ctx, span := h.o11y.Tracer().Start(ctx, "budget_update_handler.handle_invoice_item_added")
	defer span.End()

	// Parse payload
	var event struct {
		UserID         string `json:"user_id"`
		CategoryID     string `json:"category_id"`
		Amount         int64  `json:"amount"` // in cents
		Currency       string `json:"currency"`
		ReferenceMonth string `json:"reference_month"` // "2024-01"
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		h.o11y.Logger().Error(ctx, "failed to unmarshal invoice event",
			observability.Error(err),
		)
		return fmt.Errorf("failed to unmarshal invoice event: %w", err)
	}

	// Parse UUIDs
	userID, err := vos.NewUUIDFromString(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	categoryID, err := vos.NewUUIDFromString(event.CategoryID)
	if err != nil {
		return fmt.Errorf("invalid category_id: %w", err)
	}

	// Parse reference month
	referenceMonth, err := budgetVos.NewReferenceMonth(event.ReferenceMonth)
	if err != nil {
		return fmt.Errorf("invalid reference_month: %w", err)
	}

	// Parse currency
	currency, err := vos.NewCurrency(event.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	// Create Money from cents
	amount, err := vos.NewMoney(event.Amount, currency)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Increment spent amount
	if err := h.incrementSpentUseCase.Execute(ctx, userID, referenceMonth, categoryID, amount); err != nil {
		h.o11y.Logger().Error(ctx, "failed to increment spent amount",
			observability.Error(err),
			observability.String("user_id", event.UserID),
			observability.String("category_id", event.CategoryID),
		)
		return err
	}

	h.o11y.Logger().Info(ctx, "invoice event processed",
		observability.String("user_id", event.UserID),
		observability.String("category_id", event.CategoryID),
		observability.Int64("amount_cents", event.Amount),
	)

	return nil
}
