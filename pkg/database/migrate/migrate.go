package migrate

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jailtonjunior94/financial/pkg/logger"
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
		logger  logger.Logger
		migrate *migrate.Migrate
	}
)

func (m *migration) Execute() error {
	_, _, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		m.logger.Info(err.Error())
		return ErrMigrateVersion
	}

	err = m.migrate.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		m.logger.Info(err.Error())
		return nil
	}

	if err != nil {
		m.logger.Error(err.Error())
		return err
	}
	return nil
}
