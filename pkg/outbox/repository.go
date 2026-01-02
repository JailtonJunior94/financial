package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository define operações de persistência para eventos outbox.
type Repository interface {
	Save(ctx context.Context, event *OutboxEvent) error
	FindPendingBatch(ctx context.Context, limit int) ([]*OutboxEvent, error)
	UpdateStatus(ctx context.Context, event *OutboxEvent) error
	DeleteOldPublished(ctx context.Context, olderThan time.Duration) (int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*OutboxEvent, error)
	CountByStatus(ctx context.Context, status OutboxStatus) (int64, error)
}
