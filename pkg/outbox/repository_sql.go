package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/google/uuid"
)

type sqlRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

// NewRepository cria uma nova instância do repository SQL.
func NewRepository(
	db database.DBTX,
	o11y observability.Observability,
) Repository {
	return &sqlRepository{
		db:   db,
		o11y: o11y,
	}
}

// Save persiste um evento outbox dentro de uma transação.
func (r *sqlRepository) Save(ctx context.Context, event *OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (
			id, aggregate_id, aggregate_type, event_type,
			payload, status, retry_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		event.Status,
		event.RetryCount,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

// FindPendingBatch busca eventos pendentes usando SELECT FOR UPDATE SKIP LOCKED.
// Esta query evita lock contention em ambientes concorrentes.
func (r *sqlRepository) FindPendingBatch(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	query := `
		SELECT 
			id, aggregate_id, aggregate_type, event_type,
			payload, status, retry_count, published_at, failed_at, created_at
		FROM outbox_events
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.QueryContext(ctx, query, StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.o11y.Logger().Error(ctx, "failed to close rows in FindPendingBatch",
				observability.Error(closeErr),
			)
		}
	}()

	events := make([]*OutboxEvent, 0, limit)
	for rows.Next() {
		event := &OutboxEvent{}
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Payload,
			&event.Status,
			&event.RetryCount,
			&event.PublishedAt,
			&event.FailedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating outbox events: %w", err)
	}

	return events, nil
}

// UpdateStatus atualiza o status e timestamps relacionados de um evento.
func (r *sqlRepository) UpdateStatus(ctx context.Context, event *OutboxEvent) error {
	query := `
		UPDATE outbox_events
		SET status = $1,
		    retry_count = $2,
		    published_at = $3,
		    failed_at = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		event.Status,
		event.RetryCount,
		event.PublishedAt,
		event.FailedAt,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update outbox event status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

// DeleteOldPublished remove eventos publicados há mais de `olderThan`.
// Usa batch delete para performance em grandes volumes.
func (r *sqlRepository) DeleteOldPublished(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffDate := time.Now().Add(-olderThan)

	query := `
		DELETE FROM outbox_events
		WHERE status = $1
		  AND published_at IS NOT NULL
		  AND published_at < $2
	`

	result, err := r.db.ExecContext(ctx, query, StatusPublished, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old published events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// FindByID busca um evento por ID.
func (r *sqlRepository) FindByID(ctx context.Context, id uuid.UUID) (*OutboxEvent, error) {
	query := `
		SELECT 
			id, aggregate_id, aggregate_type, event_type,
			payload, status, retry_count, published_at, failed_at, created_at
		FROM outbox_events
		WHERE id = $1
	`

	event := &OutboxEvent{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.AggregateID,
		&event.AggregateType,
		&event.EventType,
		&event.Payload,
		&event.Status,
		&event.RetryCount,
		&event.PublishedAt,
		&event.FailedAt,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrEventNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find outbox event by id: %w", err)
	}

	return event, nil
}

// CountByStatus retorna a quantidade de eventos por status.
func (r *sqlRepository) CountByStatus(ctx context.Context, status OutboxStatus) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM outbox_events
		WHERE status = $1
	`

	var count int64
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events by status: %w", err)
	}

	return count, nil
}
