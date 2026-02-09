package outbox

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/google/uuid"
)

// ProcessedEventsRepository gerencia eventos já processados para idempotência.
type ProcessedEventsRepository interface {
	IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error)
	MarkAsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) error
}

type processedEventsRepository struct {
	db database.DBTX
}

// NewProcessedEventsRepository cria uma nova instância do repository.
func NewProcessedEventsRepository(db database.DBTX) ProcessedEventsRepository {
	return &processedEventsRepository{
		db: db,
	}
}

// IsProcessed verifica se o evento já foi processado pelo consumer.
func (r *processedEventsRepository) IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error) {
	query := `
		SELECT 1
		FROM processed_events
		WHERE event_id = $1 AND consumer_name = $2
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRowContext(ctx, query, eventID, consumerName).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check if event is processed: %w", err)
	}

	return true, nil
}

// MarkAsProcessed marca o evento como processado pelo consumer.
// IMPORTANTE: Deve ser chamado dentro da mesma transação do processamento.
func (r *processedEventsRepository) MarkAsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) error {
	query := `
		INSERT INTO processed_events (event_id, consumer_name, processed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (event_id, consumer_name) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, eventID, consumerName)
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}
