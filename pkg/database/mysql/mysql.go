package mysql

import (
	"database/sql"
	"errors"

	"github.com/jailtonjunior94/financial/configs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrSQLOpenConn = errors.New("unable to open connection with SQL database")
)

func NewMySqlDatabase(config *configs.Config) (*sql.DB, error) {
	sqlDB, err := sql.Open("mysql", "root:financial@tcp(localhost:3306)/financial")
	if err != nil {
		return nil, ErrSQLOpenConn
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, ErrSQLOpenConn
	}
	return sqlDB, nil
}
