package migrate

import (
	"errors"

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
	version, _, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return ErrMigrateVersion
	}

	err = m.migrate.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		if forceErr := m.migrate.Force(int(version)); forceErr != nil {
			return forceErr
		}
	}
	return nil
}
