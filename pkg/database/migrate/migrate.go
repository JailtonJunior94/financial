package migrate

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jailtonjunior94/financial/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ErrMigrateVersion        = errors.New("error checking migration version")
	ErrPostgresMigrateDriver = errors.New("unable to instantiate postgres migration driver")
)

type Migrate struct {
	logger  logger.Logger
	migrate *migrate.Migrate
}

func NewMigrate(logger logger.Logger, db *sql.DB, migratePath, dbName string) (*Migrate, error) {
	if db == nil {
		return nil, ErrPostgresMigrateDriver
	}

	driver, err := postgresmigrate.WithInstance(db, &postgresmigrate.Config{})
	if err != nil {
		return nil, ErrPostgresMigrateDriver
	}

	m, err := migrate.NewWithDatabaseInstance(migratePath, dbName, driver)
	if err != nil {
		return nil, err
	}
	return &Migrate{logger: logger, migrate: m}, nil
}

func (m *Migrate) ExecuteMigration() error {
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
