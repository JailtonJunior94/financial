package migrate

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	mysqlMigrate "github.com/golang-migrate/migrate/v4/database/mysql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrateMySql(logger logger.Logger, db *sql.DB, migratePath, dbName string) (Migrate, error) {
	if db == nil {
		return nil, ErrDatabaseConnection
	}

	driver, err := mysqlMigrate.WithInstance(db, &mysqlMigrate.Config{})
	if err != nil {
		return nil, ErrUnableToCreateDriver
	}

	migrateInstance, err := migrate.NewWithDatabaseInstance(migratePath, dbName, driver)
	if err != nil {
		return nil, err
	}
	return &migration{logger: logger, migrate: migrateInstance}, nil
}
