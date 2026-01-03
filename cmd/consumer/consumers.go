package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/lifecycle"
	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/jailtonjunior94/financial/pkg/brokers/rabbitmq"

	// Domain imports
	budgetUseCase "github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	budgetConsumers "github.com/jailtonjunior94/financial/internal/budget/infrastructure/rabbitmq"

	// Devkit imports
	"github.com/JailtonJunior94/devkit-go/pkg/consumer"
	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	devkitRabbit "github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"
)

// Run inicia o consumer server com lifecycle management.
func Run() error {
	// Load configuration
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup Observability
	o11yConfig := &otel.Config{
		Environment:     cfg.Environment,
		ServiceName:     cfg.ConsumerConfig.ServiceName,
		ServiceVersion:  cfg.O11yConfig.ServiceVersion,
		OTLPEndpoint:    cfg.O11yConfig.ExporterEndpoint,
		OTLPProtocol:    otel.OTLPProtocol(cfg.O11yConfig.ExporterProtocol),
		Insecure:        cfg.O11yConfig.ExporterInsecure,
		TraceSampleRate: cfg.O11yConfig.TraceSampleRate,
		LogLevel:        observability.LogLevel(cfg.O11yConfig.LogLevel),
		LogFormat:       observability.LogFormat(cfg.O11yConfig.LogFormat),
	}

	o11y, err := otel.NewProvider(ctx, o11yConfig)
	if err != nil {
		return fmt.Errorf("failed to create observability provider: %v", err)
	}
	defer shutdownComponent(o11y, "observability")

	o11y.Logger().Info(ctx, "starting consumer server",
		observability.String("service", cfg.ConsumerConfig.ServiceName),
		observability.String("environment", cfg.Environment),
		observability.String("broker_type", cfg.ConsumerConfig.BrokerType),
	)

	// Setup Database
	dbManager, err := postgres.New(
		cfg.DBConfig.DSN(),
		postgres.WithConnMaxLifetime(5*time.Minute),
		postgres.WithConnMaxIdleTime(2*time.Minute),
		postgres.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
		postgres.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer shutdownComponent(dbManager, "database")

	o11y.Logger().Info(ctx, "database connection established")

	// Create consumer factory based on broker type
	consumerFactory, err := createConsumerFactory(ctx, cfg, o11y)
	if err != nil {
		return fmt.Errorf("failed to create consumer factory: %v", err)
	}
	defer consumerFactory.Shutdown(ctx)

	// Setup Use Cases
	unitOfWork := uow.NewUnitOfWork(dbManager.DB())
	incrementSpentUseCase := budgetUseCase.NewIncrementSpentAmountUseCase(unitOfWork, o11y)

	// Create Domain Handlers
	budgetHandler := createBudgetHandler(incrementSpentUseCase, o11y)

	// Build Consumers using factory
	budgetConsumer, err := consumerFactory.BuildBudgetConsumer(ctx, budgetHandler)
	if err != nil {
		return fmt.Errorf("failed to build budget consumer: %v", err)
	}

	o11y.Logger().Info(ctx, "consumers created successfully")

	// Create Lifecycle Group
	serviceGroup := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())

	// Register consumers as services
	serviceGroup.Register(budgetConsumer)

	o11y.Logger().Info(ctx, "services registered in lifecycle group")

	// Start all services
	if err := serviceGroup.Start(ctx); err != nil {
		return fmt.Errorf("failed to start services: %v", err)
	}

	o11y.Logger().Info(ctx, "consumer server started successfully",
		observability.Int("worker_count", cfg.ConsumerConfig.WorkerCount),
		observability.Int("prefetch_count", cfg.ConsumerConfig.PrefetchCount),
	)

	// Wait for shutdown signal
	<-ctx.Done()

	o11y.Logger().Info(context.Background(), "shutdown signal received, initiating graceful shutdown...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := serviceGroup.Shutdown(shutdownCtx); err != nil {
		o11y.Logger().Error(context.Background(), "error during shutdown", observability.Error(err))
		return err
	}

	o11y.Logger().Info(context.Background(), "consumer server stopped gracefully")
	return nil
}

// ConsumerFactory abstrai criação de consumers por broker.
type ConsumerFactory interface {
	BuildBudgetConsumer(ctx context.Context, handler messaging.Handler) (lifecycle.Service, error)
	Shutdown(ctx context.Context) error
}

// createConsumerFactory factory method para criar factory baseado em config.
func createConsumerFactory(ctx context.Context, cfg *configs.Config, o11y observability.Observability) (ConsumerFactory, error) {
	switch cfg.ConsumerConfig.BrokerType {
	case "rabbitmq":
		rabbitClient, err := devkitRabbit.New(
			o11y,
			devkitRabbit.WithCloudConnection(cfg.RabbitMQConfig.URL),
			devkitRabbit.WithAutoReconnect(true),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create rabbitmq client: %w", err)
		}

		// Declare exchange
		if err := rabbitClient.DeclareExchange(
			ctx,
			cfg.ConsumerConfig.Exchange,
			"topic",
			true,  // durable
			false, // auto-delete
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to declare exchange: %w", err)
		}

		o11y.Logger().Info(ctx, "rabbitmq initialized",
			observability.String("exchange", cfg.ConsumerConfig.Exchange),
		)

		return &rabbitmqFactory{
			client: rabbitClient,
			cfg:    cfg,
			o11y:   o11y,
		}, nil

	case "kafka":
		return nil, fmt.Errorf("kafka not implemented yet")

	case "sqs":
		return nil, fmt.Errorf("sqs not implemented yet")

	default:
		return nil, fmt.Errorf("unknown broker type: %s", cfg.ConsumerConfig.BrokerType)
	}
}

// rabbitmqFactory cria consumers RabbitMQ.
type rabbitmqFactory struct {
	client *devkitRabbit.Client
	cfg    *configs.Config
	o11y   observability.Observability
}

// BuildBudgetConsumer cria consumer para eventos de budget.
func (f *rabbitmqFactory) BuildBudgetConsumer(ctx context.Context, handler messaging.Handler) (lifecycle.Service, error) {
	// Configuração do consumer
	consumerConfig := &rabbitmq.ConsumerConfig{
		QueueName:     "budget.updates",
		Exchange:      f.cfg.ConsumerConfig.Exchange,
		RoutingKeys: []string{
			"transaction.transaction_created",
			"invoice.invoice_item_added",
		},
		WorkerCount:   f.cfg.ConsumerConfig.WorkerCount,
		PrefetchCount: f.cfg.ConsumerConfig.PrefetchCount,
		Durable:       true,
		AutoDelete:    false,
	}

	// Build consumer usando builder pattern
	builder := rabbitmq.NewBuilder(f.client, f.o11y)
	consumer, err := builder.BuildConsumer(ctx, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build consumer: %w", err)
	}

	// Register handler
	if err := consumer.RegisterHandler(handler); err != nil {
		return nil, fmt.Errorf("failed to register handler: %w", err)
	}

	f.o11y.Logger().Info(ctx, "budget consumer built successfully",
		observability.String("queue", consumerConfig.QueueName),
		observability.Int("routing_keys", len(consumerConfig.RoutingKeys)),
	)

	// Adapt to lifecycle.Service
	return rabbitmq.NewConsumerService(consumer), nil
}

// Shutdown fecha recursos do factory.
func (f *rabbitmqFactory) Shutdown(ctx context.Context) error {
	return f.client.Shutdown(ctx)
}

// createBudgetHandler cria handler de budget adaptado para messaging.Handler.
func createBudgetHandler(
	incrementSpentUseCase budgetUseCase.IncrementSpentAmountUseCase,
	o11y observability.Observability,
) messaging.Handler {
	// Usa o handler existente do domínio
	budgetHandlerFunc := budgetConsumers.NewBudgetUpdateHandler(incrementSpentUseCase, o11y)

	// Adapta para messaging.Handler
	return messaging.NewFuncHandler(
		[]string{
			"transaction.transaction_created",
			"invoice.invoice_item_added",
		},
		func(ctx context.Context, msg *messaging.Message) error {
			// Converte messaging.Message para devkit-go consumer.Message
			devkitMsg := &consumer.Message{
				Topic:   msg.Topic,
				Value:   msg.Payload,
				Headers: convertToStringMap(msg.Headers),
			}

			// Chama handler do domínio
			return budgetHandlerFunc(ctx, devkitMsg)
		},
	)
}

// convertToStringMap converte map[string]interface{} para map[string]string.
func convertToStringMap(headers map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

// shutdownComponent helper para shutdown de componentes.
func shutdownComponent(component interface{ Shutdown(context.Context) error }, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := component.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down %s: %v\n", name, err)
	}
}
