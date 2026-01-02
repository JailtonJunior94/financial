package outbox

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/jobs"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// DispatcherJob implementa jobs.Job para dispatch de eventos outbox.
type DispatcherJob struct {
	dispatcher Dispatcher
	schedule   string
	o11y       observability.Observability
}

// NewDispatcherJob cria um novo job de dispatcher.
func NewDispatcherJob(
	dispatcher Dispatcher,
	schedule string,
	o11y observability.Observability,
) jobs.Job {
	return &DispatcherJob{
		dispatcher: dispatcher,
		schedule:   schedule,
		o11y:       o11y,
	}
}

// Name retorna o identificador do job.
func (j *DispatcherJob) Name() string {
	return "outbox_dispatcher"
}

// Schedule retorna a expressão cron para agendamento.
// Padrão: "@every 5s" - executa a cada 5 segundos para baixa latência.
func (j *DispatcherJob) Schedule() string {
	if j.schedule != "" {
		return j.schedule
	}
	return "@every 5s"
}

// Run executa o dispatch de eventos pendentes.
func (j *DispatcherJob) Run(ctx context.Context) error {
	ctx, span := j.o11y.Tracer().Start(ctx, "outbox.dispatcher_job.run")
	defer span.End()

	processed, err := j.dispatcher.Dispatch(ctx)
	if err != nil {
		j.o11y.Logger().Error(ctx, "outbox dispatcher job failed",
			observability.Error(err),
			observability.Int("processed", processed),
		)
		return fmt.Errorf("outbox dispatcher job: %w", err)
	}

	if processed > 0 {
		j.o11y.Logger().Info(ctx, "outbox dispatcher job completed",
			observability.Int("processed", processed),
		)
	}

	return nil
}

// CleanupJob implementa jobs.Job para limpeza de eventos antigos.
type CleanupJob struct {
	cleaner  Cleaner
	schedule string
	o11y     observability.Observability
}

// NewCleanupJob cria um novo job de cleanup.
func NewCleanupJob(
	cleaner Cleaner,
	schedule string,
	o11y observability.Observability,
) jobs.Job {
	return &CleanupJob{
		cleaner:  cleaner,
		schedule: schedule,
		o11y:     o11y,
	}
}

// Name retorna o identificador do job.
func (j *CleanupJob) Name() string {
	return "outbox_cleanup"
}

// Schedule retorna a expressão cron para agendamento.
// Padrão: "0 2 * * *" - executa diariamente às 2h da manhã.
func (j *CleanupJob) Schedule() string {
	if j.schedule != "" {
		return j.schedule
	}
	return "0 2 * * *"
}

// Run executa a limpeza de eventos antigos.
func (j *CleanupJob) Run(ctx context.Context) error {
	ctx, span := j.o11y.Tracer().Start(ctx, "outbox.cleanup_job.run")
	defer span.End()

	deleted, err := j.cleaner.Cleanup(ctx)
	if err != nil {
		j.o11y.Logger().Error(ctx, "outbox cleanup job failed",
			observability.Error(err),
		)
		return fmt.Errorf("outbox cleanup job: %w", err)
	}

	j.o11y.Logger().Info(ctx, "outbox cleanup job completed",
		observability.Int64("deleted", deleted),
	)

	return nil
}
