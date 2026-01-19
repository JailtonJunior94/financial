package events

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// TransactionCreatedEvent é emitido quando um novo item de transação é criado.
//
// Este evento é processado por:
// - Budget Consumer: atualiza spent_amount da categoria correspondente
//
// Payload inclui todos os campos necessários para atualização de budget:
// - transaction_id: ID do item de transação criado
// - user_id: ID do usuário dono da transação
// - category_id: ID da categoria (usado para atualizar budget)
// - amount: Valor da transação (usado para calcular spent_amount)
// - direction: Direção (income/expense) - budget ignora income
// - type: Tipo da transação (regular/credit_card/transfer)
// - reference_month: Mês de referência (formato: "2024-01").
type TransactionCreatedEvent struct{
	eventID        sharedVos.UUID
	transactionID  sharedVos.UUID
	userID         sharedVos.UUID
	categoryID     sharedVos.UUID
	amount         sharedVos.Money
	direction      transactionVos.TransactionDirection
	transactionType transactionVos.TransactionType
	referenceMonth transactionVos.ReferenceMonth
	occurredAt     time.Time
}

// NewTransactionCreatedEvent cria um novo evento TransactionCreated.
func NewTransactionCreatedEvent(
	transactionID sharedVos.UUID,
	userID sharedVos.UUID,
	categoryID sharedVos.UUID,
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	transactionType transactionVos.TransactionType,
	referenceMonth transactionVos.ReferenceMonth,
) (*TransactionCreatedEvent, error) {
	eventID, err := sharedVos.NewUUID()
	if err != nil {
		return nil, err
	}

	return &TransactionCreatedEvent{
		eventID:         eventID,
		transactionID:   transactionID,
		userID:          userID,
		categoryID:      categoryID,
		amount:          amount,
		direction:       direction,
		transactionType: transactionType,
		referenceMonth:  referenceMonth,
		occurredAt:      time.Now().UTC(),
	}, nil
}

// ID retorna o ID único do evento.
func (e *TransactionCreatedEvent) ID() sharedVos.UUID {
	return e.eventID
}

// Type retorna o tipo do evento.
func (e *TransactionCreatedEvent) Type() string {
	return "transaction_created"
}

// AggregateID retorna o ID da transação (aggregate).
func (e *TransactionCreatedEvent) AggregateID() sharedVos.UUID {
	return e.transactionID
}

// AggregateType retorna o tipo do aggregate.
func (e *TransactionCreatedEvent) AggregateType() string {
	return "transaction"
}

// OccurredAt retorna quando o evento ocorreu.
func (e *TransactionCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Payload retorna os dados do evento como map para serialização JSON.
func (e *TransactionCreatedEvent) Payload() map[string]any {
	return map[string]any{
		"transaction_id":  e.transactionID.String(),
		"user_id":         e.userID.String(),
		"category_id":     e.categoryID.String(),
		"amount":          e.amount.Cents(),
		"currency":        e.amount.Currency().String(),
		"direction":       e.direction.String(),
		"type":            e.transactionType.String(),
		"reference_month": e.referenceMonth.String(),
	}
}
