package events

import (
	"time"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/events"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	InvoiceEvent struct {
		EventName   vos.InvoiceEvent `json:"event_name"`
		Payload     any              `json:"payload"`
		OccurrredAt time.Time        `json:"occurred_at"`
	}

	InvoiceCreatedEvent struct {
		InvoiceID sharedVos.UUID `json:"invoice_id"`
		UserID    sharedVos.UUID `json:"user_id"`
	}
)

func NewInvoiceCreatedEvent(invoiceID, userID sharedVos.UUID) events.Event {
	return InvoiceEvent{
		EventName:   vos.InvoiceCreated,
		Payload:     InvoiceCreatedEvent{InvoiceID: invoiceID, UserID: userID},
		OccurrredAt: time.Now(),
	}
}

func (e InvoiceEvent) GetEventType() string {
	return e.EventName.String()
}

func (e InvoiceEvent) GetPayload() any {
	return e.Payload
}
