package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type CleanupConfig struct {
	// RetentionDays dias de retenção para eventos publicados com sucesso.
	RetentionDays int

	// FailedRetentionDays dias de retenção para eventos com falha permanente.
	// Valor menor que RetentionDays para liberar espaço mas manter histórico para análise.
	FailedRetentionDays int

	// ProcessedEventsRetentionDays dias de retenção para registros de idempotência.
	// Deve ser maior que o tempo máximo de redelivery do broker para evitar re-processamento.
	ProcessedEventsRetentionDays int

	DryRun bool
}

func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		RetentionDays:                90,
		FailedRetentionDays:          30,
		ProcessedEventsRetentionDays: 30,
		DryRun:                       false,
	}
}

type Cleaner interface {
	Cleanup(ctx context.Context) (int64, error)
}

type cleaner struct {
	uow    uow.UnitOfWork
	config *CleanupConfig
	o11y   observability.Observability
}

func NewCleaner(
	uow uow.UnitOfWork,
	config *CleanupConfig,
	o11y observability.Observability,
) Cleaner {
	return &cleaner{
		uow:    uow,
		config: config,
		o11y:   o11y,
	}
}

func (c *cleaner) Cleanup(ctx context.Context) (int64, error) {
	ctx, span := c.o11y.Tracer().Start(ctx, "outbox.cleaner.cleanup")
	defer span.End()

	publishedRetention := time.Duration(c.config.RetentionDays) * 24 * time.Hour
	failedRetention := time.Duration(c.config.FailedRetentionDays) * 24 * time.Hour
	processedEventsRetention := time.Duration(c.config.ProcessedEventsRetentionDays) * 24 * time.Hour

	c.o11y.Logger().Info(
		ctx,
		"starting outbox cleanup",
		observability.Int("retention_days", c.config.RetentionDays),
		observability.Int("failed_retention_days", c.config.FailedRetentionDays),
		observability.Int("processed_events_retention_days", c.config.ProcessedEventsRetentionDays),
		observability.Bool("dry_run", c.config.DryRun),
	)

	var deletedPublished, deletedFailed, deletedProcessed int64

	err := c.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		outboxRepo := NewRepository(tx, c.o11y)
		processedEventsRepo := NewProcessedEventsRepository(tx)

		var err error

		deletedPublished, err = outboxRepo.DeleteOldPublished(ctx, publishedRetention)
		if err != nil {
			return fmt.Errorf("delete old published: %w", err)
		}

		deletedFailed, err = outboxRepo.DeleteOldFailed(ctx, failedRetention)
		if err != nil {
			return fmt.Errorf("delete old failed: %w", err)
		}

		deletedProcessed, err = processedEventsRepo.DeleteOldProcessed(ctx, processedEventsRetention)
		if err != nil {
			return fmt.Errorf("delete old processed events: %w", err)
		}

		return nil
	})

	if err != nil {
		c.o11y.Logger().Error(ctx, "cleanup failed", observability.Error(err))
		return 0, fmt.Errorf("cleanup: %w", err)
	}

	totalDeleted := deletedPublished + deletedFailed + deletedProcessed

	c.o11y.Logger().Info(ctx, "cleanup completed",
		observability.Int64("deleted_published", deletedPublished),
		observability.Int64("deleted_failed", deletedFailed),
		observability.Int64("deleted_processed", deletedProcessed),
		observability.Int64("total_deleted", totalDeleted),
	)

	return totalDeleted, nil
}
