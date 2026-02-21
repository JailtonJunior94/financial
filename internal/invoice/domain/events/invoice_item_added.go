// NOTA: InvoiceItemAddedEvent faz parte de uma futura pipeline de eventos de
// granularidade fina (evento por item adicionado à fatura).
// Atualmente NÃO está conectado ao outbox nem ao Budget Consumer.
// O Budget Consumer usa PurchaseEventPayload (via outbox/RabbitMQ), não este evento.
// O comentário abaixo descreve a intenção de design para quando este evento for integrado.
// Não remover — é parte do vocabulário de domínio previsto para evolução futura.
package events

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceItemAddedEvent é emitido quando um novo item é adicionado a uma fatura.
//
// Quando integrado ao outbox, será processado por:
// - Budget Consumer: atualiza spent_amount da categoria correspondente
//
// Payload inclui todos os campos necessários para atualização de budget:
// - invoice_item_id: ID do item de fatura criado
// - invoice_id: ID da fatura
// - user_id: ID do usuário dono da fatura
// - category_id: ID da categoria (usado para atualizar budget)
// - amount: Valor da parcela (usado para calcular spent_amount)
// - reference_month: Mês de referência da fatura (formato: "2024-01").
type InvoiceItemAddedEvent struct {
	eventID        sharedVos.UUID
	invoiceItemID  sharedVos.UUID
	invoiceID      sharedVos.UUID
	userID         sharedVos.UUID
	categoryID     sharedVos.UUID
	amount         sharedVos.Money
	referenceMonth pkgVos.ReferenceMonth
	occurredAt     time.Time
}

// NewInvoiceItemAddedEvent cria um novo evento InvoiceItemAdded.
func NewInvoiceItemAddedEvent(
	invoiceItemID sharedVos.UUID,
	invoiceID sharedVos.UUID,
	userID sharedVos.UUID,
	categoryID sharedVos.UUID,
	amount sharedVos.Money,
	referenceMonth pkgVos.ReferenceMonth,
) (*InvoiceItemAddedEvent, error) {
	eventID, err := sharedVos.NewUUID()
	if err != nil {
		return nil, err
	}

	return &InvoiceItemAddedEvent{
		eventID:        eventID,
		invoiceItemID:  invoiceItemID,
		invoiceID:      invoiceID,
		userID:         userID,
		categoryID:     categoryID,
		amount:         amount,
		referenceMonth: referenceMonth,
		occurredAt:     time.Now().UTC(),
	}, nil
}

// ID retorna o ID único do evento.
func (e *InvoiceItemAddedEvent) ID() sharedVos.UUID {
	return e.eventID
}

// Type retorna o tipo do evento.
func (e *InvoiceItemAddedEvent) Type() string {
	return "invoice_item_added"
}

// AggregateID retorna o ID da fatura (aggregate root).
func (e *InvoiceItemAddedEvent) AggregateID() sharedVos.UUID {
	return e.invoiceID
}

// AggregateType retorna o tipo do aggregate.
func (e *InvoiceItemAddedEvent) AggregateType() string {
	return "invoice"
}

// OccurredAt retorna quando o evento ocorreu.
func (e *InvoiceItemAddedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// InvoiceItemAddedSchemaVersion é a versão do schema do payload deste evento.
// Incrementar quando o contrato mudar de forma incompatível.
const InvoiceItemAddedSchemaVersion = 1

// Payload retorna os dados do evento como map para serialização JSON.
func (e *InvoiceItemAddedEvent) Payload() map[string]any {
	return map[string]any{
		"version":         InvoiceItemAddedSchemaVersion,
		"invoice_item_id": e.invoiceItemID.String(),
		"invoice_id":      e.invoiceID.String(),
		"user_id":         e.userID.String(),
		"category_id":     e.categoryID.String(),
		"amount":          e.amount.Cents(),
		"currency":        e.amount.Currency().String(),
		"reference_month": e.referenceMonth.String(),
	}
}
