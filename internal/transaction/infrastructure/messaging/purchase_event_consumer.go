package messaging

import (
	"context"

	"github.com/jailtonjunior94/financial/pkg/messaging"
)

// PurchaseEventConsumer handles purchase-related events (stub — replaced in Task 10).
type PurchaseEventConsumer struct{}

// NewPurchaseEventConsumer creates a PurchaseEventConsumer stub.
func NewPurchaseEventConsumer() *PurchaseEventConsumer {
	return &PurchaseEventConsumer{}
}

// Topics returns the event topics this consumer handles.
func (c *PurchaseEventConsumer) Topics() []string {
	return []string{}
}

// Handle processes an incoming message.
func (c *PurchaseEventConsumer) Handle(_ context.Context, _ *messaging.Message) error {
	return nil
}
