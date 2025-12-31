package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/budget"
	"github.com/jailtonjunior94/financial/internal/card"
	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/invoice"
	"github.com/jailtonjunior94/financial/internal/payment_method"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/JailtonJunior94/devkit-go/pkg/database/postgres"
	httpserver "github.com/JailtonJunior94/devkit-go/pkg/http_server/chi_server"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"
)

func Run() error {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("run: failed to load config: %v", err)
	}

	log.Printf("starting financial api on port %s (environment: %s)", cfg.HTTPConfig.Port, cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	o11yConfig := &otel.Config{
		Environment:     cfg.Environment,
		ServiceName:     cfg.O11yConfig.ServiceName,
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
		return fmt.Errorf("run: failed to create observability provider: %v", err)
	}

	o11y.Logger().Info(context.Background(), "o11y initialized", observability.String("service", cfg.O11yConfig.ServiceName))

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
		return fmt.Errorf("run: failed to connect to database: %v", err)
	}

	userModule := user.NewUserModule(dbManager.DB(), cfg, o11y)

	// JWT adapter para validação de tokens (usado por múltiplos módulos)
	jwtAdapter := auth.NewJwtAdapter(cfg, o11y)

	categoryModule := category.NewCategoryModule(dbManager.DB(), o11y, jwtAdapter)
	cardModule := card.NewCardModule(dbManager.DB(), o11y, jwtAdapter)
	paymentMethodModule := payment_method.NewPaymentMethodModule(dbManager.DB(), o11y)
	budgetModule := budget.NewBudgetModule(dbManager.DB(), o11y)

	// ✅ Invoice module depends on CardProvider from card module
	invoiceModule := invoice.NewInvoiceModule(dbManager.DB(), o11y, jwtAdapter, cardModule.CardProvider)

	srv := httpserver.New(
		o11y,
		httpserver.WithMetrics(),
		httpserver.WithCORS("*"),
		httpserver.WithPort(cfg.HTTPConfig.Port),
		httpserver.WithServiceName(cfg.O11yConfig.ServiceName),
		httpserver.WithServiceVersion(cfg.O11yConfig.ServiceVersion),
		httpserver.WithHealthChecks(map[string]httpserver.HealthCheckFunc{"database": dbManager.Ping}),
	)

	srv.RegisterRouters(userModule.UserRouter)
	srv.RegisterRouters(categoryModule.CategoryRouter)
	srv.RegisterRouters(cardModule.CardRouter)
	srv.RegisterRouters(paymentMethodModule.PaymentMethodRouter)
	srv.RegisterRouters(budgetModule.BudgetRouter)
	srv.RegisterRouters(invoiceModule.InvoiceRouter)

	go func() {
		<-ctx.Done()
		o11y.Logger().Info(context.Background(), "shutting down gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := o11y.Shutdown(shutdownCtx); err != nil {
			o11y.Logger().Error(context.Background(), "error during o11y shutdown", observability.Error(err))
		}

		if err := dbManager.Shutdown(shutdownCtx); err != nil {
			o11y.Logger().Error(context.Background(), "error during database shutdown", observability.Error(err))
		}
	}()

	return srv.Start(ctx)
}
