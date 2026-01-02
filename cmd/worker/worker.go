package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/worker/jobs"
	pkgjobs "github.com/jailtonjunior94/financial/pkg/jobs"
	"github.com/jailtonjunior94/financial/pkg/scheduler"

	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres"
	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"
)

// Run inicia o worker de cron jobs.
// Inicializa todas as dependências necessárias (DB, RabbitMQ, Observability),
// registra os jobs configurados e inicia o scheduler.
// Implementa graceful shutdown aguardando jobs em execução finalizarem.
func Run() error {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("worker: failed to load config: %v", err)
	}

	// Context principal com captura de sinais de shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Inicializar Observability (OpenTelemetry)
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

	o11y.Logger().Info(ctx, "initializing worker",
		observability.String("service", cfg.WorkerConfig.ServiceName),
		observability.String("environment", cfg.Environment),
	)

	// Inicializar Database Manager (reutiliza conexão existente)
	dbManager, err := postgres.New(
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBConfig.Host,
			cfg.DBConfig.Port,
			cfg.DBConfig.User,
			cfg.DBConfig.Password,
			cfg.DBConfig.Name,
		),
		postgres.WithConnMaxLifetime(5*time.Minute),
		postgres.WithConnMaxIdleTime(2*time.Minute),
		postgres.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
		postgres.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
	)
	if err != nil {
		return fmt.Errorf("worker: failed to connect to database: %v", err)
	}

	o11y.Logger().Info(ctx, "database connection established")

	// Inicializar RabbitMQ Client (reutiliza configuração existente)
	rabbitClient, err := rabbitmq.New(
		o11y,
		rabbitmq.WithCloudConnection(cfg.RabbitMQConfig.URL),
		rabbitmq.WithPublisherConfirms(true),
		rabbitmq.WithAutoReconnect(true),
	)
	if err != nil {
		return fmt.Errorf("worker: failed to create rabbitmq client: %v", err)
	}

	// Declarar exchange (idempotente - não recria se já existir)
	if err := rabbitClient.DeclareExchange(ctx, cfg.RabbitMQConfig.Exchange, "topic", true, false, nil); err != nil {
		return fmt.Errorf("worker: failed to declare exchange: %v", err)
	}

	o11y.Logger().Info(ctx, "rabbitmq initialized",
		observability.String("exchange", cfg.RabbitMQConfig.Exchange),
		observability.String("url", cfg.RabbitMQConfig.URL),
	)

	// Configuração de jobs (timeout, recovery, concorrência)
	jobConfig := &pkgjobs.Config{
		DefaultTimeout:    time.Duration(cfg.WorkerConfig.DefaultTimeoutSeconds) * time.Second,
		EnableRecovery:    true,
		MaxConcurrentJobs: cfg.WorkerConfig.MaxConcurrentJobs,
	}

	// Criar scheduler
	sched := scheduler.New(ctx, o11y, jobConfig)

	// Registrar jobs
	// IMPORTANTE: Adicione seus jobs customizados aqui
	jobsToRegister := []pkgjobs.Job{
		// Exemplo 1: Job que usa banco de dados
		jobs.NewDatabaseCleanupJob(dbManager.DB(), o11y),

		// Exemplo 2: Job que publica mensagens no RabbitMQ
		jobs.NewReportGeneratorJob(dbManager.DB(), rabbitClient, cfg.RabbitMQConfig.Exchange, o11y),
	}

	for _, job := range jobsToRegister {
		if err := sched.Register(job); err != nil {
			return fmt.Errorf("worker: failed to register job %s: %v", job.Name(), err)
		}
	}

	// Iniciar scheduler
	sched.Start()

	o11y.Logger().Info(ctx, "worker started successfully",
		observability.Int("jobs_registered", len(jobsToRegister)),
	)

	// Aguardar sinal de shutdown
	<-ctx.Done()

	o11y.Logger().Info(context.Background(), "shutdown signal received, initiating graceful shutdown...")

	// Graceful shutdown com timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Parar scheduler (aguarda jobs em execução)
	if err := sched.Shutdown(shutdownCtx); err != nil {
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
