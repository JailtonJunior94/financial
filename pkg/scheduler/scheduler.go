package scheduler

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/jailtonjunior94/financial/pkg/jobs"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/robfig/cron/v3"
)

// Scheduler gerencia o ciclo de vida de cron jobs.
// Implementa graceful shutdown e controle de execução concorrente.
type Scheduler struct {
	cron   *cron.Cron
	o11y   observability.Observability
	config *jobs.Config
	jobs   []jobs.Job

	// Controle de lifecycle
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	runningMux sync.Mutex
	running    map[string]int // rastreia jobs em execução por nome
}

// New cria uma nova instância do scheduler.
// Usa as opções fornecidas para configurar comportamento de recovery e logging.
func New(ctx context.Context, o11y observability.Observability, config *jobs.Config) *Scheduler {
	if config == nil {
		config = jobs.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(ctx)

	// Configurar cron com recovery se habilitado
	opts := []cron.Option{
		cron.WithSeconds(), // Suporte para jobs com precisão de segundos
	}

	if config.EnableRecovery {
		opts = append(opts, cron.WithChain(
			cron.Recover(cron.DefaultLogger), // Recovery nativo do cron
		))
	}

	return &Scheduler{
		cron:    cron.New(opts...),
		o11y:    o11y,
		config:  config,
		jobs:    make([]jobs.Job, 0),
		ctx:     ctx,
		cancel:  cancel,
		running: make(map[string]int),
	}
}

// Register registra um job no scheduler.
// O job será agendado de acordo com sua expressão Schedule().
// Retorna error se a expressão cron for inválida.
func (s *Scheduler) Register(job jobs.Job) error {
	s.jobs = append(s.jobs, job)

	// Wrapper que adiciona logging, timeout e controle de concorrência
	wrappedFunc := s.wrapJob(job)

	_, err := s.cron.AddFunc(job.Schedule(), wrappedFunc)
	if err != nil {
		return fmt.Errorf("failed to register job %s: %w", job.Name(), err)
	}

	s.o11y.Logger().Info(
		s.ctx,
		fmt.Sprintf("job registered: %s", job.Name()),
		observability.String("schedule", job.Schedule()),
		observability.String("job", job.Name()),
	)

	return nil
}

// Start inicia o scheduler.
// Jobs serão executados de acordo com seus agendamentos.
// Não bloqueia - retorna imediatamente após iniciar.
func (s *Scheduler) Start() {
	s.o11y.Logger().Info(
		s.ctx,
		"starting scheduler",
		observability.Int("jobs_count", len(s.jobs)),
		observability.Int64("default_timeout_seconds", int64(s.config.DefaultTimeout.Seconds())),
	)

	s.cron.Start()
}

// Shutdown para o scheduler gracefully.
// Aguarda jobs em execução finalizarem ou timeout expirar.
// Retorna error se contexto expirar antes de todos os jobs finalizarem.
func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.o11y.Logger().Info(ctx, "shutting down scheduler...")

	// Sinaliza para jobs pararem de executar
	s.cancel()

	// Para de aceitar novos jobs
	cronCtx := s.cron.Stop()

	// Aguarda término de jobs agendados pelo cron
	select {
	case <-cronCtx.Done():
		s.o11y.Logger().Info(ctx, "cron scheduler stopped")
	case <-ctx.Done():
		s.o11y.Logger().Warn(ctx, "cron shutdown timeout exceeded")
	}

	// Aguarda jobs em execução finalizarem
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.o11y.Logger().Info(ctx, "all running jobs completed")
		return nil
	case <-ctx.Done():
		s.runningMux.Lock()
		runningCount := 0
		for _, count := range s.running {
			runningCount += count
		}
		s.runningMux.Unlock()

		s.o11y.Logger().Warn(
			ctx,
			"shutdown timeout exceeded with jobs still running",
			observability.Int("running_jobs", runningCount),
		)
		return fmt.Errorf("shutdown timeout exceeded with %d jobs still running", runningCount)
	}
}

// wrapJob encapsula um job com logging, timeout, recovery e controle de concorrência.
func (s *Scheduler) wrapJob(job jobs.Job) func() {
	return func() {
		jobName := job.Name()

		// Verifica se deve executar (scheduler ainda ativo)
		select {
		case <-s.ctx.Done():
			s.o11y.Logger().Info(
				context.Background(),
				"skipping job execution (scheduler shutting down)",
				observability.String("job", jobName),
			)
			return
		default:
		}

		// Controle de concorrência (opcional)
		if s.config.MaxConcurrentJobs > 0 {
			s.runningMux.Lock()
			currentRunning := s.running[jobName]
			if currentRunning >= s.config.MaxConcurrentJobs {
				s.runningMux.Unlock()
				s.o11y.Logger().Warn(
					s.ctx,
					"job skipped (max concurrent executions reached)",
					observability.String("job", jobName),
					observability.Int("max_concurrent", s.config.MaxConcurrentJobs),
				)
				return
			}
			s.running[jobName]++
			s.runningMux.Unlock()

			defer func() {
				s.runningMux.Lock()
				s.running[jobName]--
				s.runningMux.Unlock()
			}()
		}

		// Incrementa wait group
		s.wg.Add(1)
		defer s.wg.Done()

		// Recovery adicional (além do cron recovery) para logging detalhado
		defer func() {
			if r := recover(); r != nil {
				s.o11y.Logger().Error(context.Background(),
					"job panic recovered",
					observability.String("job", jobName),
					observability.Any("panic", r),
					observability.String("stack", string(debug.Stack())),
				)
			}
		}()

		// Contexto com timeout para execução do job
		ctx, cancel := context.WithTimeout(s.ctx, s.config.DefaultTimeout)
		defer cancel()

		start := time.Now()
		s.o11y.Logger().Info(ctx,
			"job started",
			observability.String("job", jobName),
		)

		// Executa o job
		err := job.Run(ctx)

		duration := time.Since(start)

		if err != nil {
			s.o11y.Logger().Error(ctx,
				"job failed",
				observability.String("job", jobName),
				observability.Int64("duration_ms", duration.Milliseconds()),
				observability.Error(err),
			)
			return
		}

		s.o11y.Logger().Info(ctx,
			"job completed successfully",
			observability.String("job", jobName),
			observability.Int64("duration_ms", duration.Milliseconds()),
		)
	}
}
