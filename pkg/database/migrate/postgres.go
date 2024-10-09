package migrate

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"

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
