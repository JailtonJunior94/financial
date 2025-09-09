package database

import (
	"context"
	"database/sql"
)

type DBExecutor interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type dbExecutor struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *dbExecutor {
	return &dbExecutor{db: db}
}
