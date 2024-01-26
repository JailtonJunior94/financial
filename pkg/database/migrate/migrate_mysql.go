package migrate

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	mysqlMigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jailtonjunior94/financial/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrateMySql(logger logger.Logger, db *sql.DB, migratePath, dbName string) (*Migrate, error) {
	if db == nil {
		return nil, ErrPostgresMigrateDriver
	}

	driver, err := mysqlMigrate.WithInstance(db, &mysqlMigrate.Config{})
	if err != nil {
		return nil, ErrPostgresMigrateDriver
	}

	m, err := migrate.NewWithDatabaseInstance(migratePath, dbName, driver)
	if err != nil {
		return nil, err
	}
	return &Migrate{logger: logger, migrate: m}, nil
}

func (m *Migrate) ExecuteMigrationMySql() error {
	_, _, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		m.logger.Error(err.Error())
		return ErrMigrateVersion
	}

	err = m.migrate.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		m.logger.Error(err.Error())
		return nil
	}

	if err != nil {
		m.logger.Error(err.Error())
		return err
	}
	return nil
}
