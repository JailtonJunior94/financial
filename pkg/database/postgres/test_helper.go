package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	migration "github.com/jailtonjunior94/financial/pkg/database/migrate"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
}

func NewPostgresContainer(ctx context.Context) *PostgresContainer {
	postgres, err := postgres.Run(
		ctx,
		"postgres:16.2-alpine",
		postgres.WithDatabase("financial"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	if err != nil {
		log.Fatalf("could not start postgres container: %s", err)
	}
	return &PostgresContainer{Container: postgres}
}

func (s *PostgresContainer) ExecuteMigration(db *sql.DB, dbName, migratePath string) {
	migrate, err := migration.NewMigratePostgres(db, migratePath, dbName)
	if err != nil {
		log.Fatalf("could not start migrate: %s", err)
	}

	if err = migrate.Execute(); err != nil {
		log.Fatalf("could execute migration: %s", err)
	}
}
