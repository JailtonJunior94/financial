package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/financial/configs"

	_ "github.com/lib/pq"
)

var (
	ErrSQLOpenConn = errors.New("unable to open connection with SQL database")
)

func NewPostgresDatabase(config *configs.Config) (*sql.DB, error) {
	sqlDB, err := sql.Open(config.DBConfig.Driver, dsn(config))
	if err != nil {
		return nil, ErrSQLOpenConn
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, ErrSQLOpenConn
	}
	sqlDB.SetMaxIdleConns(config.DBConfig.DBMaxIdleConns)
	return sqlDB, nil
}

func dsn(config *configs.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBConfig.Host,
		config.DBConfig.Port,
		config.DBConfig.User,
		config.DBConfig.Password,
		config.DBConfig.Name,
	)
}
