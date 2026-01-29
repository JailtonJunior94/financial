package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres_otelsql"
	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"
)

type application struct {
	config    *configs.Config
	wg        sync.WaitGroup
	dbManager *postgres_otelsql.DBManager
	ctx       context.Context
	cancel    context.CancelFunc
	client    *rabbitmq.Client
	consumer  *rabbitmq.Consumer
	o11y      observability.Observability
}

func Run() error {
	app, err := NewApplication()
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.o11y.Logger().Info(ctx, "consumer is running")

	<-ctx.Done()

	app.o11y.Logger().Info(ctx, "shutdown signal received")

	if err := app.Stop(30 * time.Second); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	return nil
}

func NewApplication() (*application, error) {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	otelConfig := &otel.Config{
		Environment:     cfg.Environment,
		ServiceVersion:  cfg.O11yConfig.ServiceVersion,
		ServiceName:     cfg.ConsumerConfig.ServiceName,
		OTLPEndpoint:    cfg.O11yConfig.ExporterEndpoint,
		TraceSampleRate: cfg.O11yConfig.TraceSampleRate,
		Insecure:        cfg.O11yConfig.ExporterInsecure,
		LogLevel:        observability.LogLevel(cfg.O11yConfig.LogLevel),
		LogFormat:       observability.LogFormat(cfg.O11yConfig.LogFormat),
		OTLPProtocol:    otel.OTLPProtocol(cfg.O11yConfig.ExporterProtocol),
	}

	o11y, err := otel.NewProvider(context.Background(), otelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup observability: %w", err)
	}

	o11y.Logger().Info(
		context.Background(),
		"observability initialized",
		observability.String("service", cfg.ConsumerConfig.ServiceName),
		observability.String("version", cfg.O11yConfig.ServiceVersion),
	)

	dbManager, err := database.NewDatabaseManager(
		context.Background(),
		database.WithMetrics(true),
		database.WithDSN(cfg.DBConfig.DSN()),
		database.WithConnMaxLifetime(5*time.Minute),
		database.WithConnMaxIdleTime(2*time.Minute),
		database.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
		database.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
		database.WithServiceName(cfg.ConsumerConfig.ServiceName),
	)
	if err != nil {
		return nil, fmt.Errorf("run: failed to connect to database: %v", err)
	}

	client, err := rabbitmq.New(
		o11y,
		rabbitmq.WithCloudConnection(cfg.RabbitMQConfig.URL),
		rabbitmq.WithServiceName(cfg.ConsumerConfig.ServiceName),
		rabbitmq.WithServiceVersion(cfg.O11yConfig.ServiceVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbitmq client: %w", err)
	}

	consumer := rabbitmq.NewConsumer(
		client,
		rabbitmq.WithWorkerPool(10),
		rabbitmq.WithPrefetchCount(10),
		rabbitmq.WithQueue(cfg.RabbitMQConfig.Queue),
	)

	ctx, cancel := context.WithCancel(context.Background())

	return &application{
		config:    cfg,
		ctx:       ctx,
		o11y:      o11y,
		cancel:    cancel,
		client:    client,
		consumer:  consumer,
		dbManager: dbManager,
	}, nil
}

func (app *application) Start() error {
	app.o11y.Logger().Info(app.ctx, "starting consumer...")

	app.wg.Go(func() {
		if err := app.consumer.Start(app.ctx); err != nil {
			app.o11y.Logger().Error(app.ctx, "consumer stopped with error", observability.Error(err))
		}
	})

	app.o11y.Logger().Info(app.ctx, "consumer started successfully")
	return nil
}

func (app *application) Stop(timeout time.Duration) error {
	app.o11y.Logger().Info(app.ctx, "stopping consumer...")

	// 1. Cancelar context para sinalizar goroutines pararem
	app.cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	// 2. Aguardar goroutines finalizarem (processando mensagens em andamento)
	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.o11y.Logger().Info(shutdownCtx, "all goroutines stopped")
	case <-shutdownCtx.Done():
		app.o11y.Logger().Warn(shutdownCtx, "shutdown timeout exceeded")
		return fmt.Errorf("shutdown timeout exceeded")
	}

	// 3. Fechar consumer (UMA ÚNICA VEZ, após goroutines pararem)
	if err := app.consumer.Close(); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing consumer", observability.Error(err))
	}

	// 4. Fechar RabbitMQ client
	if err := app.client.Shutdown(shutdownCtx); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing rabbitmq client", observability.Error(err))
	}

	// 5. Fechar conexão com banco de dados
	if err := app.dbManager.Shutdown(shutdownCtx); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing database connection", observability.Error(err))
	}

	app.o11y.Logger().Info(shutdownCtx, "consumer stopped successfully")
	return nil
}
