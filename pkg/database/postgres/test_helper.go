package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	migration "github.com/jailtonjunior94/financial/pkg/database/migrate"

	"github.com/JailtonJunior94/devkit-go/pkg/logger"

	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container *postgresContainer.PostgresContainer
}

func NewPostgresContainer(ctx context.Context) *PostgresContainer {
	postgres, err := postgresContainer.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:16.2-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
		postgresContainer.WithDatabase("financial"),
		postgresContainer.WithUsername("postgres"),
		postgresContainer.WithPassword("postgres"),
	)

	if err != nil {
		log.Fatalf("could not start postgres container: %s", err)
	}
	return &PostgresContainer{Container: postgres}
}

func (s *PostgresContainer) ExecuteMigration(logger logger.Logger, db *sql.DB, dbName, migratePath string) {
	migrate, err := migration.NewMigratePostgres(logger, db, migratePath, dbName)
	if err != nil {
		log.Fatalf("could not start migrate: %s", err)
	}

	if err = migrate.Execute(); err != nil {
		log.Fatalf("could execute migration: %s", err)
	}
}
