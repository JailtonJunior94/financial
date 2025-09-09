package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/database"
)

type UnitOfWork interface {
	Executor() database.DBExecutor
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type unitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Executor() database.DBExecutor {
	if u.tx != nil {
		return u.tx
	}
	return u.db
}

func (u *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	u.tx = tx

	if err = fn(ctx); err != nil {
		if errRollback := u.Rollback(); errRollback != nil {
			return fmt.Errorf("original error: %s, rollback error: %s", err, errRollback)
		}
		return err
	}
	return u.CommitOrRollback()
}

func (u *unitOfWork) CommitOrRollback() error {
	if err := u.tx.Commit(); err != nil {
		if errRollback := u.Rollback(); errRollback != nil {
			return fmt.Errorf("original error: %s, rollback error: %s", err, errRollback)
		}
		return err
	}
	u.tx = nil
	return nil
}

func (u *unitOfWork) Rollback() error {
	if u.tx != nil {
		return errors.New("no transaction to rollback")
	}

	if err := u.tx.Rollback(); err != nil {
		return err
	}

	u.tx = nil
	return nil
}
