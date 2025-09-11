package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
)

var (
	ErrMigrateVersion       = errors.New("error on migrate version")
	ErrDatabaseConnection   = errors.New("database connection is nil")
	ErrUnableToCreateDriver = errors.New("unable to create driver instance")
)

type (
	Migrate interface {
		Execute() error
	}

	migration struct {
		migrate *migrate.Migrate
	}
)

func (m *migration) Execute() error {
	_, _, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return ErrMigrateVersion
	}

	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: %v: %v", ErrMigrateVersion, err)
	}
	return nil
}
