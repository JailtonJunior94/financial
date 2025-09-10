package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jailtonjunior94/financial/pkg/database/migrate"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

type CockroachDBContainer struct {
	container *cockroachdb.CockroachDBContainer
}

func SetupCockroachDB(ctx context.Context, t testing.TB) *CockroachDBContainer {
	cockroachDB, err := cockroachdb.Run(
		ctx,
		"cockroachdb/cockroach:v23.2.4",
		cockroachdb.WithInsecure(),
		cockroachdb.WithUser("dbUser"),
		cockroachdb.WithDatabase("financial"),
	)
	if err != nil {
		t.Fatalf("failed to start CockroachDB container: %v", err)
	}

	state, err := cockroachDB.State(ctx)
	if err != nil {
		t.Fatalf("failed to get CockroachDB container state: %v", err)
	}

	if !state.Running {
		t.Fatalf("CockroachDB container is not running")
	}

	return &CockroachDBContainer{container: cockroachDB}
}

func (c *CockroachDBContainer) Teardown(ctx context.Context, t testing.TB) {
	if err := c.container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate CockroachDB container: %v", err)
	}
}

func (c *CockroachDBContainer) DSN(ctx context.Context, t testing.TB) string {
	cfg, err := c.container.ConnectionConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get CockroachDB connection config: %v", err)
	}
	return cfg.ConnString()
}

func (c *CockroachDBContainer) Migrate(t testing.TB, db *sql.DB, migratePath, dbName string) {
	migration, err := migrate.NewMigrateCockroachDB(db, migratePath, dbName)
	if err != nil {
		t.Fatalf("failed to create migration instance: %v", err)
	}

	if err := migration.Execute(); err != nil {
		t.Fatalf("failed to execute migration: %v", err)
	}
}
