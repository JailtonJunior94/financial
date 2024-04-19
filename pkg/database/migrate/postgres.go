package migrate

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jailtonjunior94/financial/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigratePostgres(logger logger.Logger, db *sql.DB, migratePath, dbName string) (Migrate, error) {
	if db == nil {
		return nil, ErrDatabaseConnection
	}

	driver, err := postgresMigrate.WithInstance(db, &postgresMigrate.Config{})
	if err != nil {
		return nil, ErrUnableToCreateDriver
	}

	migrateInstance, err := migrate.NewWithDatabaseInstance(migratePath, dbName, driver)
	if err != nil {
		return nil, err
	}
	return &migration{logger: logger, migrate: migrateInstance}, nil
}
