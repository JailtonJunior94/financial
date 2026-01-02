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
	RetentionDays int
	DryRun        bool
}

func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		RetentionDays: 90,
		DryRun:        false,
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

	retentionDuration := time.Duration(c.config.RetentionDays) * 24 * time.Hour

	c.o11y.Logger().Info(
		ctx,
		"starting outbox cleanup",
		observability.Int("retention_days", c.config.RetentionDays),
		observability.Bool("dry_run", c.config.DryRun),
	)

	var deleted int64
	err := c.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		var err error
		outboxRepository := NewRepository(tx, c.o11y)

		deleted, err = outboxRepository.DeleteOldPublished(ctx, retentionDuration)
		if err != nil {
			return fmt.Errorf("delete old published: %w", err)
		}
		return nil
	})

	if err != nil {
		c.o11y.Logger().Error(ctx, "cleanup failed", observability.Error(err))
		return 0, fmt.Errorf("cleanup: %w", err)
	}

	c.o11y.Logger().Info(ctx, "cleanup completed",
		observability.Int64("deleted_count", deleted),
	)

	return deleted, nil
}
