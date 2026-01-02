package consumer

func Run() error {
	// 	cfg, err := configs.LoadConfig(".")
	// if err != nil {
	// 	log.Fatalf("failed to load config: %v", err)
	// }

	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer stop()

	// o11yConfig := &otel.Config{
	// 	Environment:     cfg.Environment,
	// 	ServiceName:     cfg.ConsumerConfig.ServiceName,
	// 	ServiceVersion:  cfg.O11yConfig.ServiceVersion,
	// 	OTLPEndpoint:    cfg.O11yConfig.ExporterEndpoint,
	// 	OTLPProtocol:    otel.OTLPProtocol(cfg.O11yConfig.ExporterProtocol),
	// 	Insecure:        cfg.O11yConfig.ExporterInsecure,
	// 	TraceSampleRate: cfg.O11yConfig.TraceSampleRate,
	// 	LogLevel:        observability.LogLevel(cfg.O11yConfig.LogLevel),
	// 	LogFormat:       observability.LogFormat(cfg.O11yConfig.LogFormat),
	// }

	// o11y, err := otel.NewProvider(ctx, o11yConfig)
	// if err != nil {
	// 	log.Fatalf("failed to create observability provider: %v", err)
	// }

	// defer func() {
	// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	// 	if err := o11y.Shutdown(shutdownCtx); err != nil {
	// 		log.Printf("error shutting down observability: %v", err)
	// 	}
	// }()

	// if err := consumer.RunConsumers(ctx, cfg, o11y); err != nil {
	// 	log.Fatalf("consumer server failed: %v", err)
	// }
	// o11y.Logger().Info(ctx, "starting consumer server",
	// 	observability.String("service", cfg.ConsumerConfig.ServiceName),
	// 	observability.String("environment", cfg.Environment),
	// )

	// dbManager, err := postgres.New(
	// 	cfg.DBConfig.DSN(),
	// 	postgres.WithConnMaxLifetime(5*time.Minute),
	// 	postgres.WithConnMaxIdleTime(2*time.Minute),
	// 	postgres.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
	// 	postgres.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to connect to database: %w", err)
	// }
	// defer func() {
	// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	// 	if err := dbManager.Shutdown(shutdownCtx); err != nil {
	// 		o11y.Logger().Error(context.Background(), "error shutting down database", observability.Error(err))
	// 	}
	// }()

	// rabbitClient, err := rabbitmq.New(
	// 	o11y,
	// 	rabbitmq.WithCloudConnection(cfg.RabbitMQConfig.URL),
	// 	rabbitmq.WithAutoReconnect(true),
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to create rabbitmq client: %w", err)
	// }
	// defer func() {
	// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	// 	if err := rabbitClient.Shutdown(shutdownCtx); err != nil {
	// 		o11y.Logger().Error(context.Background(), "error shutting down rabbitmq", observability.Error(err))
	// 	}
	// }()

	// queue, err := rabbitClient.DeclareQueue(ctx, "budget.updates", true, false, false, nil)
	// if err != nil {
	// 	return fmt.Errorf("failed to declare budget queue: %w", err)
	// }

	// if err := rabbitClient.BindQueue(ctx, queue.Name, "transaction.transaction_created", cfg.RabbitMQConfig.Exchange, nil); err != nil {
	// 	return fmt.Errorf("failed to bind transaction events: %w", err)
	// }

	// if err := rabbitClient.BindQueue(ctx, queue.Name, "invoice.invoice_item_added", cfg.RabbitMQConfig.Exchange, nil); err != nil {
	// 	return fmt.Errorf("failed to bind invoice events: %w", err)
	// }

	// consumerServer := consumer.New(
	// 	o11y,
	// 	consumer.WithEnvironment(cfg.Environment),
	// 	consumer.WithServiceName(cfg.ConsumerConfig.ServiceName),
	// 	consumer.WithServiceVersion(cfg.O11yConfig.ServiceVersion),
	// 	consumer.WithWorkerCount(5), // 5 concurrent workers
	// 	consumer.WithShutdownTimeout(30*time.Second),
	// )

	// o11y.Logger().Info(ctx, "consumer server created with idempotency middleware")

	// unitOfWork := uow.NewUnitOfWork(dbManager.DB())
	// incrementSpentUseCase := budgetUseCase.NewIncrementSpentAmountUseCase(unitOfWork, o11y)
	// budgetHandler := budgetConsumers.NewBudgetUpdateHandler(incrementSpentUseCase, o11y)

	// consumerServer.RegisterHandlers(
	// 	consumer.NewFuncHandler(
	// 		[]string{"transaction.transaction_created", "invoice.invoice_item_added"},
	// 		budgetHandler,
	// 	),
	// )

	// o11y.Logger().Info(ctx, "budget update handler registered")

	// o11y.Logger().Info(ctx, "consumer server configured",
	// 	observability.Int("concurrency", 5),
	// 	observability.String("graceful_shutdown_timeout", "30s"),
	// )

	// o11y.Logger().Info(ctx, "consumer server starting...")
	// if err := consumerServer.Start(ctx); err != nil {
	// 	return fmt.Errorf("consumer server error: %w", err)
	// }

	// o11y.Logger().Info(ctx, "consumer server stopped gracefully")
	return nil
}
