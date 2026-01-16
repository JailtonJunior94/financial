package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	pkgjobs "github.com/jailtonjunior94/financial/pkg/jobs"
	"github.com/jailtonjunior94/financial/pkg/outbox"
	"github.com/jailtonjunior94/financial/pkg/scheduler"

	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"
)

func Run() error {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("worker: failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	o11yConfig := &otel.Config{
		Environment:     cfg.Environment,
		ServiceName:     cfg.WorkerConfig.ServiceName,
		ServiceVersion:  cfg.O11yConfig.ServiceVersion,
		OTLPEndpoint:    cfg.O11yConfig.ExporterEndpoint,
		OTLPProtocol:    otel.OTLPProtocol(cfg.O11yConfig.ExporterProtocol),
		Insecure:        cfg.O11yConfig.ExporterInsecure,
		TraceSampleRate: cfg.O11yConfig.TraceSampleRate,
		LogLevel:        observability.LogLevel(cfg.O11yConfig.LogLevel),
		LogFormat:       observability.LogFormat(cfg.O11yConfig.LogFormat),
	}

	o11y, err := otel.NewProvider(context.Background(), o11yConfig)
	if err != nil {
		return fmt.Errorf("worker: failed to create observability provider: %v", err)
	}

	dbManager, err := postgres.New(
		cfg.DBConfig.DSN(),
		postgres.WithConnMaxLifetime(5*time.Minute),
		postgres.WithConnMaxIdleTime(2*time.Minute),
		postgres.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
		postgres.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
	)
	if err != nil {
		return fmt.Errorf("worker: failed to connect to database: %v", err)
	}
	o11y.Logger().Info(ctx, "database connection established")

	rabbitClient, err := rabbitmq.New(
		o11y,
		rabbitmq.WithAutoReconnect(true),
		rabbitmq.WithPublisherConfirms(true),
		rabbitmq.WithCloudConnection(cfg.RabbitMQConfig.URL),
	)
	if err != nil {
		return fmt.Errorf("worker: failed to create rabbitmq client: %v", err)
	}

	if err := rabbitClient.DeclareExchange(ctx, cfg.RabbitMQConfig.Exchange, "topic", true, false, nil); err != nil {
		return fmt.Errorf("worker: failed to declare exchange: %v", err)
	}

	o11y.Logger().Info(ctx, "rabbitmq initialized",
		observability.String("exchange", cfg.RabbitMQConfig.Exchange),
		observability.String("url", cfg.RabbitMQConfig.URL),
	)

	uow, err := uow.NewUnitOfWork(dbManager.DB())
	if err != nil {
		return fmt.Errorf("worker: failed to create unit of work: %v", err)
	}
	outboxDispatcher := outbox.NewDispatcher(uow, rabbitClient, outbox.DefaultDispatcherConfig(cfg.RabbitMQConfig.Exchange), o11y)
	outboxCleanup := outbox.NewCleaner(uow, outbox.DefaultCleanupConfig(), o11y)

	jobsToRegister := []pkgjobs.Job{
		outbox.NewDispatcherJob(outboxDispatcher, "@every 5s", o11y),
		outbox.NewCleanupJob(outboxCleanup, "@daily", o11y),
	}

	scheduler := scheduler.New(ctx, o11y, pkgjobs.DefaultConfig())

	for _, job := range jobsToRegister {
		if err := scheduler.Register(job); err != nil {
			return fmt.Errorf("worker: failed to register job %s: %v", job.Name(), err)
		}
	}

	scheduler.Start()

	o11y.Logger().Info(ctx, "worker started successfully", observability.Int("jobs_registered", len(jobsToRegister)))

	// Aguardar sinal de shutdown
	<-ctx.Done()

	o11y.Logger().Info(context.Background(), "shutdown signal received, initiating graceful shutdown...")

	// Graceful shutdown com timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Parar scheduler (aguarda jobs em execução)
	if err := scheduler.Shutdown(shutdownCtx); err != nil {
		o11y.Logger().Error(context.Background(), "error during scheduler shutdown", observability.Error(err))
	}

	// 2. Fechar RabbitMQ
	if err := rabbitClient.Shutdown(shutdownCtx); err != nil {
		o11y.Logger().Error(context.Background(), "error during rabbitmq shutdown", observability.Error(err))
	}

	// 3. Fechar Database
	if err := dbManager.Shutdown(shutdownCtx); err != nil {
		o11y.Logger().Error(context.Background(), "error during database shutdown", observability.Error(err))
	}

	// 4. Fechar Observability
	if err := o11y.Shutdown(shutdownCtx); err != nil {
		o11y.Logger().Error(context.Background(), "error during o11y shutdown", observability.Error(err))
	}

	o11y.Logger().Info(context.Background(), "worker shutdown completed")
	return nil
}
