package consumer

import (
	"context"
	"errors"
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
	"github.com/jailtonjunior94/financial/internal/transaction"
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

	// ✅ Declarar exchange principal
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

	// ✅ Configurar Dead Letter Queue quando habilitado
	// mainQueueArgs receberá x-dead-letter-exchange se DLQExchange estiver configurado.
	// ATENÇÃO: alterar args de uma queue existente requer deletá-la e recriar.
	var mainQueueArgs map[string]interface{}
	if cfg.RabbitMQConfig.DLQExchange != "" {
		if err := client.DeclareExchange(
			context.Background(),
			cfg.RabbitMQConfig.DLQExchange,
			"fanout",
			true,  // durable
			false, // auto-delete
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to declare dlq exchange: %w", err)
		}

		_, err = client.DeclareQueue(
			context.Background(),
			cfg.RabbitMQConfig.DLQQueue,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare dlq queue: %w", err)
		}

		// fanout exchange: routing key vazia
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

	// ✅ Declarar queue principal
	_, err = client.DeclareQueue(
		context.Background(),
		cfg.RabbitMQConfig.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		mainQueueArgs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// ✅ Bind queue ao exchange com routing key pattern
	// Pattern "invoice.#" captura todos os eventos de invoice
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

	jwtAdapter := auth.NewJwtAdapter(app.config, app.o11y)

	fm := metrics.NewFinancialMetrics(app.o11y)
	invoiceRepo := invoicerepos.NewInvoiceRepository(app.dbManager.DB(), app.o11y, fm)
	invoiceTotalProvider := invoiceadapters.NewInvoiceTotalProviderAdapter(invoiceRepo)
	invoiceCategoryTotalProvider := invoiceadapters.NewInvoiceCategoryTotalAdapter(invoiceRepo)

	transactionModule, err := transaction.NewTransactionModule(
		app.dbManager.DB(),
		app.o11y,
		jwtAdapter,
		invoiceTotalProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction module: %w", err)
	}

	budgetModule, err := budget.NewBudgetModule(
		app.dbManager.DB(),
		app.o11y,
		jwtAdapter,
		invoiceCategoryTotalProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create budget module: %w", err)
	}

	// Registrar handler composto: transaction sync + budget sync para cada evento de purchase.
	// Cada consumer mantém idempotência independente via consumerName distinto.
	//
	// Ambos os handlers executam independentemente: se um falhar, o outro ainda roda.
	// errors.Join agrega os erros para que a mensagem seja reenfileirada (NACK) apenas
	// se algum handler falhou, sem impedir que o outro seja executado.
	// Na reentrega, cada handler verifica sua própria idempotência via processed_events,
	// saltando o processamento que já foi concluído com sucesso.
	topics := transactionModule.PurchaseEventConsumer.Topics()
	for _, topic := range topics {
		handler := func(ctx context.Context, msg rabbitmq.Message) error {
			m := &messaging.Message{
				ID:      msg.MessageID,
				Topic:   msg.RoutingKey,
				Payload: msg.Body,
				Headers: msg.Headers,
			}

			var errs []error

			if err := transactionModule.PurchaseEventConsumer.Handle(ctx, m); err != nil {
				errs = append(errs, fmt.Errorf("transaction: %w", err))
			}

			if err := budgetModule.BudgetEventConsumer.Handle(ctx, m); err != nil {
				errs = append(errs, fmt.Errorf("budget: %w", err))
			}

			return errors.Join(errs...)
		}
		app.consumer.RegisterHandler(topic, handler)
	}

	app.o11y.Logger().Info(app.ctx, "handlers registered",
		observability.Int("handlers_count", len(topics)),
		observability.Any("topics", topics),
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
