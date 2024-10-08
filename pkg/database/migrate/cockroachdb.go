package migrate

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	cockroachdbMigrate "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/jailtonjunior94/financial/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrateCockroachDB(logger logger.Logger, db *sql.DB, migratePath, dbName string) (Migrate, error) {
	if db == nil {
		return nil, ErrDatabaseConnection
	}

	driver, err := cockroachdbMigrate.WithInstance(db, &cockroachdbMigrate.Config{})
	if err != nil {
		return nil, ErrUnableToCreateDriver
	}

	migrateInstance, err := migrate.NewWithDatabaseInstance(migratePath, dbName, driver)
	if err != nil {
		return nil, err
	}
	return &migration{logger: logger, migrate: migrateInstance}, nil
}
