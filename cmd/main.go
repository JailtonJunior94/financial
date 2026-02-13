// Package main é o ponto de entrada da Financial API.
//
//	@title			Financial API
//	@version		1.0.0
//	@description	API para gestão financeira pessoal com suporte a cartões de crédito, transações,
//	@description	orçamentos mensais, categorias e faturas com parcelamento.
//	@description
//	@description	## Valores Monetários
//	@description	Todos os valores monetários são representados como **strings decimais** (ex: `"1234.56"`)
//	@description	para preservar precisão e evitar erros de ponto flutuante.
//	@description
//	@description	## Paginação
//	@description	Endpoints de listagem usam **cursor-based pagination**. Parâmetros:
//	@description	- `limit` (int, default: 20, max: 100)
//	@description	- `cursor` (string, opcional): valor de `pagination.next_cursor` da resposta anterior
//	@description
//	@description	## Autenticação
//	@description	Endpoints protegidos requerem o header `Authorization: Bearer {token}`.
//	@description	Obtenha o token em `POST /api/v1/token`.
//
//	@contact.name	Jailton Junior
//	@contact.url	https://github.com/jailtonjunior94
//
//	@license.name	MIT
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Token JWT no formato: `Bearer {token}`
package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/jailtonjunior94/financial/cmd/consumer"
	"github.com/jailtonjunior94/financial/cmd/server"
	"github.com/jailtonjunior94/financial/cmd/worker"
	"github.com/jailtonjunior94/financial/configs"

	"github.com/JailtonJunior94/devkit-go/pkg/migration"

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

	consumer := &cobra.Command{
		Use:   "consumer",
		Short: "Financial Consumer",
		Run: func(cmd *cobra.Command, args []string) {
			if err := consumer.Run(); err != nil {
				log.Fatalf("consumer failed: %v", err)
			}
		},
	}

	worker := &cobra.Command{
		Use:   "worker",
		Short: "Financial Worker (Cron Jobs)",
		Run: func(cmd *cobra.Command, args []string) {
			if err := worker.Run(); err != nil {
				log.Fatalf("worker failed: %v", err)
			}
		},
	}

	root.AddCommand(migrate, api, consumer, worker)
	if err := root.Execute(); err != nil {
		log.Fatalf("error executing command: %v", err)
	}
}
