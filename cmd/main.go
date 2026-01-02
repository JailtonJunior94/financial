package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/cmd/consumer"
	"github.com/jailtonjunior94/financial/cmd/server"
	"github.com/jailtonjunior94/financial/configs"

	"github.com/JailtonJunior94/devkit-go/pkg/migration"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "financial",
		Short: "Financial",
	}

	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Financial Migrations",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := configs.LoadConfig(".")
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}

			ctx := context.Background()
			logger := migration.NewSlogTextLogger(slog.LevelInfo)

			migrator, err := migration.New(
				migration.WithDriver(migration.DriverCockroachDB),
				migration.WithDSN(cfg.DBConfig.DSN()),
				migration.WithSource(cfg.DBConfig.MigratePath),
				migration.WithLogger(logger),
				migration.WithTimeout(5*time.Minute),
			)
			if err != nil {
				log.Fatalf("failed to create migrator: %v", err)
			}

			defer func() {
				if err := migrator.Close(); err != nil {
					log.Printf("Warning: failed to close migrator: %v", err)
				}
			}()

			if err := migrator.Up(ctx); err != nil {
				if migration.IsNoChangeError(err) {
					log.Println("No migrations to apply - database is up to date")
					return
				}
				log.Fatalf("failed to apply migrations: %v", err)
				return
			}

			version, dirty, err := migrator.Version(ctx)
			if err != nil {
				log.Fatalf("failed to get version: %v", err)
			}

			log.Printf("Migration completed successfully! Current version: %d (dirty: %v)", version, dirty)
		},
	}

	api := &cobra.Command{
		Use:   "api",
		Short: "Financial API",
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Run(); err != nil {
				log.Fatalf("server failed: %v", err)
			}
		},
	}

	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Financial Consumers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := configs.LoadConfig(".")
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

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
				log.Fatalf("failed to create observability provider: %v", err)
			}

			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := o11y.Shutdown(shutdownCtx); err != nil {
					log.Printf("error shutting down observability: %v", err)
				}
			}()

			if err := consumer.RunConsumers(ctx, cfg, o11y); err != nil {
				log.Fatalf("consumer server failed: %v", err)
			}
		},
	}

	root.AddCommand(migrate, api, consumers)
	if err := root.Execute(); err != nil {
		log.Fatalf("error executing command: %v", err)
	}
}
