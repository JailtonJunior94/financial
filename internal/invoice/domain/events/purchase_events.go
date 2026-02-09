package events

import "time"

const (
	// PurchaseCreatedEventName é o nome do evento de criação de purchase.
	// O aggregate_type "invoice" será prefixado automaticamente pelo dispatcher.
	// Routing key final: "invoice.purchase.created"
	PurchaseCreatedEventName = "purchase.created"
	// PurchaseUpdatedEventName é o nome do evento de atualização de purchase.
	// Routing key final: "invoice.purchase.updated"
	PurchaseUpdatedEventName = "purchase.updated"
	// PurchaseDeletedEventName é o nome do evento de exclusão de purchase.
	// Routing key final: "invoice.purchase.deleted"
	PurchaseDeletedEventName = "purchase.deleted"
)

// PurchaseEventPayload representa o payload compartilhado dos eventos de purchase.
type PurchaseEventPayload struct {
	UserID         string    `json:"user_id"`
	CategoryID     string    `json:"category_id"`
	AffectedMonths []string  `json:"affected_months"` // Meses impactados (YYYY-MM)
	OccurredAt     time.Time `json:"occurred_at"`
}

// PurchaseCreated é emitido quando uma purchase é criada.
// Implementa github.com/JailtonJunior94/devkit-go/pkg/events.Event.
type PurchaseCreated struct {
	eventType string
	payload   PurchaseEventPayload
}

// NewPurchaseCreated cria um novo evento de purchase criada.
func NewPurchaseCreated(userID, categoryID string, affectedMonths []string) *PurchaseCreated {
	return &PurchaseCreated{
		eventType: PurchaseCreatedEventName,
		payload: PurchaseEventPayload{
			UserID:         userID,
			CategoryID:     categoryID,
			AffectedMonths: affectedMonths,
			OccurredAt:     time.Now().UTC(),
		},
	}
}

// GetEventType implementa events.Event.
func (e *PurchaseCreated) GetEventType() string {
	return e.eventType
}

// GetPayload implementa events.Event.
func (e *PurchaseCreated) GetPayload() any {
	return e.payload
}

// PurchaseUpdated é emitido quando uma purchase é atualizada.
// Implementa github.com/JailtonJunior94/devkit-go/pkg/events.Event.
type PurchaseUpdated struct {
	eventType string
	payload   PurchaseEventPayload
}

// NewPurchaseUpdated cria um novo evento de purchase atualizada.
func NewPurchaseUpdated(userID, categoryID string, affectedMonths []string) *PurchaseUpdated {
	return &PurchaseUpdated{
		eventType: PurchaseUpdatedEventName,
		payload: PurchaseEventPayload{
			UserID:         userID,
			CategoryID:     categoryID,
			AffectedMonths: affectedMonths,
			OccurredAt:     time.Now().UTC(),
		},
	}
}

// GetEventType implementa events.Event.
func (e *PurchaseUpdated) GetEventType() string {
	return e.eventType
}

// GetPayload implementa events.Event.
func (e *PurchaseUpdated) GetPayload() any {
	return e.payload
}

// PurchaseDeleted é emitido quando uma purchase é deletada.
// Implementa github.com/JailtonJunior94/devkit-go/pkg/events.Event.
type PurchaseDeleted struct {
	eventType string
	payload   PurchaseEventPayload
}

// NewPurchaseDeleted cria um novo evento de purchase deletada.
func NewPurchaseDeleted(userID, categoryID string, affectedMonths []string) *PurchaseDeleted {
	return &PurchaseDeleted{
		eventType: PurchaseDeletedEventName,
		payload: PurchaseEventPayload{
			UserID:         userID,
			CategoryID:     categoryID,
			AffectedMonths: affectedMonths,
			OccurredAt:     time.Now().UTC(),
		},
	}
}

// GetEventType implementa events.Event.
func (e *PurchaseDeleted) GetEventType() string {
	return e.eventType
}

// GetPayload implementa events.Event.
func (e *PurchaseDeleted) GetPayload() any {
	return e.payload
}
