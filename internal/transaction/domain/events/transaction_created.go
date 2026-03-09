package events

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

const TransactionCreatedSchemaVersion = "2"

// TransactionCreatedEvent is emitted when a new transaction is created.
type TransactionCreatedEvent struct {
	transactionID      vos.UUID
	userID             vos.UUID
	categoryID         vos.UUID
	amount             vos.Money
	paymentMethod      transactionVos.PaymentMethod
	transactionDate    time.Time
	referenceMonth     pkgVos.ReferenceMonth
	invoiceID          *vos.UUID
	installmentNumber  *int
	installmentTotal   *int
	installmentGroupID *vos.UUID
}

// NewTransactionCreatedEvent creates a TransactionCreatedEvent.
func NewTransactionCreatedEvent(
	transactionID vos.UUID,
	userID vos.UUID,
	categoryID vos.UUID,
	amount vos.Money,
	paymentMethod transactionVos.PaymentMethod,
	transactionDate time.Time,
	referenceMonth pkgVos.ReferenceMonth,
	invoiceID *vos.UUID,
	installmentNumber *int,
	installmentTotal *int,
	installmentGroupID *vos.UUID,
) *TransactionCreatedEvent {
	return &TransactionCreatedEvent{
		transactionID:      transactionID,
		userID:             userID,
		categoryID:         categoryID,
		amount:             amount,
		paymentMethod:      paymentMethod,
		transactionDate:    transactionDate,
		referenceMonth:     referenceMonth,
		invoiceID:          invoiceID,
		installmentNumber:  installmentNumber,
		installmentTotal:   installmentTotal,
		installmentGroupID: installmentGroupID,
	}
}

// EventType returns the event type identifier.
func (e *TransactionCreatedEvent) EventType() string {
	return "transaction.created"
}

// IdempotencyKey returns a unique key for deduplication.
func (e *TransactionCreatedEvent) IdempotencyKey() string {
	return e.transactionID.String()
}

// Payload returns the event data as a map for outbox serialization.
func (e *TransactionCreatedEvent) Payload() map[string]any {
	payload := map[string]any{
		"version":              TransactionCreatedSchemaVersion,
		"transaction_id":       e.transactionID.String(),
		"user_id":              e.userID.String(),
		"category_id":          e.categoryID.String(),
		"amount":               e.amount.Cents(),
		"currency":             e.amount.Currency().String(),
		"payment_method":       e.paymentMethod.String(),
		"transaction_date":     e.transactionDate.Format("2006-01-02"),
		"reference_month":      e.referenceMonth.String(),
		"invoice_id":           nil,
		"installment_number":   nil,
		"installment_total":    nil,
		"installment_group_id": nil,
	}
	if e.invoiceID != nil {
		payload["invoice_id"] = e.invoiceID.String()
	}
	if e.installmentNumber != nil {
		payload["installment_number"] = *e.installmentNumber
	}
	if e.installmentTotal != nil {
		payload["installment_total"] = *e.installmentTotal
	}
	if e.installmentGroupID != nil {
		payload["installment_group_id"] = e.installmentGroupID.String()
	}
	return payload
}
