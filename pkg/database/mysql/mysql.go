package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/financial/configs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrSQLOpenConn = errors.New("unable to open connection with SQL database")
)

func NewMySqlDatabase(config *configs.Config) (*sql.DB, error) {
	sqlDB, err := sql.Open(config.DBDriver, dsn(config))
	if err != nil {
		return nil, ErrSQLOpenConn
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, ErrSQLOpenConn
	}
	sqlDB.SetMaxIdleConns(config.DBMaxIdleConns)
	return sqlDB, nil
}

func dsn(config *configs.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)
}
