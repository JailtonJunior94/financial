package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres_otelsql"
	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/budget"
	invoiceadapters "github.com/jailtonjunior94/financial/internal/invoice/infrastructure/adapters"
	invoicerepos "github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database"
	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
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
		rabbitmq.WithAutoReconnect(true),
		rabbitmq.WithCloudConnection(cfg.RabbitMQConfig.URL),
		rabbitmq.WithServiceName(cfg.ConsumerConfig.ServiceName),
		rabbitmq.WithServiceVersion(cfg.O11yConfig.ServiceVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbitmq client: %w", err)
	}

	if err := client.DeclareExchange(
		context.Background(),
		cfg.RabbitMQConfig.Exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	var mainQueueArgs map[string]interface{}
	if cfg.RabbitMQConfig.DLQExchange != "" {
		if err := client.DeclareExchange(
			context.Background(),
			cfg.RabbitMQConfig.DLQExchange,
			"fanout",
			true,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to declare dlq exchange: %w", err)
		}
		_, err = client.DeclareQueue(
			context.Background(),
			cfg.RabbitMQConfig.DLQQueue,
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare dlq queue: %w", err)
		}
		if err = client.BindQueue(
			context.Background(),
			cfg.RabbitMQConfig.DLQQueue,
			"",
			cfg.RabbitMQConfig.DLQExchange,
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to bind dlq queue: %w", err)
		}
		mainQueueArgs = map[string]interface{}{
			"x-dead-letter-exchange": cfg.RabbitMQConfig.DLQExchange,
		}
		o11y.Logger().Info(
			context.Background(),
			"dead letter queue configured",
			observability.String("dlq_exchange", cfg.RabbitMQConfig.DLQExchange),
			observability.String("dlq_queue", cfg.RabbitMQConfig.DLQQueue),
		)
	}
	_, err = client.DeclareQueue(
		context.Background(),
		cfg.RabbitMQConfig.Queue,
		true,
		false,
		false,
		mainQueueArgs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	err = client.BindQueue(
		context.Background(),
		cfg.RabbitMQConfig.Queue,
		"invoice.#",
		cfg.RabbitMQConfig.Exchange,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}
	o11y.Logger().Info(
		context.Background(),
		"rabbitmq topology configured",
		observability.String("exchange", cfg.RabbitMQConfig.Exchange),
		observability.String("queue", cfg.RabbitMQConfig.Queue),
		observability.String("routing_pattern", "invoice.#"),
		observability.Bool("dlq_enabled", cfg.RabbitMQConfig.DLQExchange != ""),
	)

	consumer := rabbitmq.NewConsumer(
		client,
		rabbitmq.WithWorkerPool(10),
		rabbitmq.WithPrefetchCount(10),
		rabbitmq.WithQueue(cfg.RabbitMQConfig.Queue),
	)

	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec

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

	jwtAdapter := auth.NewJwtAdapter(app.config, app.o11y)

	fm := metrics.NewFinancialMetrics(app.o11y)
	invoiceRepo := invoicerepos.NewInvoiceRepository(app.dbManager.DB(), app.o11y, fm)
	invoiceCategoryTotalProvider := invoiceadapters.NewInvoiceCategoryTotalAdapter(invoiceRepo)

	budgetModule, err := budget.NewBudgetModule(
		app.dbManager.DB(),
		app.o11y,
		jwtAdapter,
		invoiceCategoryTotalProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create budget module: %w", err)
	}

	var registeredTopics []string
	if budgetModule.BudgetEventConsumer != nil {
		for _, topic := range budgetModule.BudgetEventConsumer.Topics() {
			topic := topic
			handler := func(ctx context.Context, msg rabbitmq.Message) error {
				m := &messaging.Message{
					ID:      msg.MessageID,
					Topic:   msg.RoutingKey,
					Payload: msg.Body,
					Headers: msg.Headers,
				}
				return budgetModule.BudgetEventConsumer.Handle(ctx, m)
			}
			app.consumer.RegisterHandler(topic, handler)
			registeredTopics = append(registeredTopics, topic)
		}
	}

	app.o11y.Logger().Info(app.ctx, "handlers registered",
		observability.Int("handlers_count", len(registeredTopics)),
		observability.Any("topics", registeredTopics),
	)

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := app.consumer.Start(app.ctx); err != nil {
			app.o11y.Logger().Error(app.ctx, "consumer stopped with error", observability.Error(err))
		}
	}()

	app.o11y.Logger().Info(app.ctx, "consumer started successfully")
	return nil
}

func (app *application) Stop(timeout time.Duration) error {
	app.o11y.Logger().Info(app.ctx, "stopping consumer...")

	app.cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()
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
	if err := app.consumer.Close(); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing consumer", observability.Error(err))
	}
	if err := app.client.Shutdown(shutdownCtx); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing rabbitmq client", observability.Error(err))
	}
	if err := app.dbManager.Shutdown(shutdownCtx); err != nil {
		app.o11y.Logger().Error(shutdownCtx, "error closing database connection", observability.Error(err))
	}

	app.o11y.Logger().Info(shutdownCtx, "consumer stopped successfully")
	return nil
}
