package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpserver "github.com/JailtonJunior94/devkit-go/pkg/http_server/chi_server"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/budget"
	"github.com/jailtonjunior94/financial/internal/card"
	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/invoice"
	"github.com/jailtonjunior94/financial/internal/payment_method"
	"github.com/jailtonjunior94/financial/internal/transaction"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

func Run() error {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("run: failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	o11yConfig := &otel.Config{
		Environment:     cfg.Environment,
		ServiceName:     cfg.HTTPConfig.ServiceName,
		ServiceVersion:  cfg.O11yConfig.ServiceVersion,
		TraceSampleRate: cfg.O11yConfig.TraceSampleRate,
		OTLPEndpoint:    cfg.O11yConfig.ExporterEndpoint,
		Insecure:        cfg.O11yConfig.ExporterInsecure,
		LogLevel:        observability.LogLevel(cfg.O11yConfig.LogLevel),
		OTLPProtocol:    otel.OTLPProtocol(cfg.O11yConfig.ExporterProtocol),
		LogFormat:       observability.LogFormat(cfg.O11yConfig.LogFormat),
	}

	o11y, err := otel.NewProvider(context.Background(), o11yConfig)
	if err != nil {
		return fmt.Errorf("run: failed to create observability provider: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(
		ctx,
		database.WithMetrics(true),
		database.WithDSN(cfg.DBConfig.DSN()),
		database.WithConnMaxLifetime(5*time.Minute),
		database.WithConnMaxIdleTime(2*time.Minute),
		database.WithServiceName(cfg.HTTPConfig.ServiceName),
		database.WithMaxOpenConns(cfg.DBConfig.DBMaxOpenConns),
		database.WithMaxIdleConns(cfg.DBConfig.DBMaxIdleConns),
	)
	if err != nil {
		return fmt.Errorf("run: failed to connect to database: %v", err)
	}
	o11y.Logger().Info(ctx, "database connection established with OpenTelemetry instrumentation")

	metricsMiddleware := middlewares.NewMetricsMiddleware(o11y)

	jwtAdapter := auth.NewJwtAdapter(cfg, o11y)
	userModule := user.NewUserModule(dbManager.DB(), cfg, o11y)
	cardModule := card.NewCardModule(dbManager.DB(), o11y, jwtAdapter)

	budgetModule, err := budget.NewBudgetModule(dbManager.DB(), o11y, jwtAdapter)
	if err != nil {
		return fmt.Errorf("run: failed to create budget module: %v", err)
	}

	categoryModule := category.NewCategoryModule(dbManager.DB(), o11y, jwtAdapter)
	paymentMethodModule := payment_method.NewPaymentMethodModule(dbManager.DB(), o11y)

	// Create outbox service for transactional event persistence
	outboxRepository := outbox.NewRepository(dbManager.DB(), o11y)
	outboxService := outbox.NewService(outboxRepository, o11y)

	// Create transaction module first (without InvoiceTotalProvider for now)
	// This breaks the circular dependency between Invoice and Transaction
	transactionModule, err := transaction.NewTransactionModule(dbManager.DB(), o11y, jwtAdapter, nil)
	if err != nil {
		return fmt.Errorf("run: failed to create transaction module: %v", err)
	}

	// Now create invoice module with outbox service
	// Events are persisted in outbox and processed asynchronously by worker
	invoiceModule, err := invoice.NewInvoiceModule(dbManager.DB(), o11y, jwtAdapter, cardModule.CardProvider, outboxService)
	if err != nil {
		return fmt.Errorf("run: failed to create invoice module: %v", err)
	}

	// Re-create transaction module with InvoiceTotalProvider
	transactionModule, err = transaction.NewTransactionModule(dbManager.DB(), o11y, jwtAdapter, invoiceModule.InvoiceTotalProvider)
	if err != nil {
		return fmt.Errorf("run: failed to create transaction module: %v", err)
	}

	srv, err := httpserver.New(
		o11y,
		httpserver.WithMetrics(),
		httpserver.WithCORS("*"),
		httpserver.WithPort(cfg.HTTPConfig.Port),
		httpserver.WithServiceName(cfg.HTTPConfig.ServiceName),
		httpserver.WithServiceVersion(cfg.O11yConfig.ServiceVersion),
		httpserver.WithHealthChecks(map[string]httpserver.HealthCheckFunc{"database": dbManager.Ping}),
		httpserver.WithMiddleware(metricsMiddleware.Handler),
	)
	if err != nil {
		return fmt.Errorf("run: failed to create http server: %v", err)
	}

	srv.RegisterRouters(userModule.UserRouter)
	srv.RegisterRouters(categoryModule.CategoryRouter)
	srv.RegisterRouters(cardModule.CardRouter)
	srv.RegisterRouters(transactionModule.TransactionRouter)
	srv.RegisterRouters(paymentMethodModule.PaymentMethodRouter)
	srv.RegisterRouters(budgetModule.BudgetRouter)
	srv.RegisterRouters(invoiceModule.InvoiceRouter)

	go func() {
		<-ctx.Done()

		// Derive shutdown context from parent to maintain trace context
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		o11y.Logger().Info(shutdownCtx, "shutting down gracefully...")

		if err := o11y.Shutdown(shutdownCtx); err != nil {
			o11y.Logger().Error(shutdownCtx, "error during o11y shutdown", observability.Error(err))
		}

		if err := dbManager.Shutdown(shutdownCtx); err != nil {
			o11y.Logger().Error(shutdownCtx, "error during database shutdown", observability.Error(err))
		}
	}()

	return srv.Start(ctx)
}
