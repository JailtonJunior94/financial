package uow

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRepositoryNotRegistered     = errors.New("repository not registered")
	ErrRepositoryAlreadyRegistered = errors.New("repository already registered")
)

type Repository any
type RepositoryName string
type RepositoryFactory func(tx *sql.Tx) Repository

type TX interface {
	Get(name RepositoryName) (Repository, error)
}

type UnitOfWork interface {
	Register(name RepositoryName, factory RepositoryFactory) error
	Remove(name RepositoryName) error
	Has(name RepositoryName) bool
	Clear()
	Do(ctx context.Context, fn func(ctx context.Context, tx TX) error) error
}

type unitOfWork struct {
	db           *sql.DB
	repositories map[RepositoryName]RepositoryFactory
}

func NewUnitOfWork(db *sql.DB) *unitOfWork {
	return &unitOfWork{
		db:           db,
		repositories: make(map[RepositoryName]RepositoryFactory),
	}
}

func (u *unitOfWork) Register(name RepositoryName, factory RepositoryFactory) error {
	if _, ok := u.repositories[name]; ok {
		return ErrRepositoryAlreadyRegistered
	}

	u.repositories[name] = factory
	return nil
}

func (u *unitOfWork) Remove(name RepositoryName) error {
	if _, ok := u.repositories[name]; !ok {
		return ErrRepositoryNotRegistered
	}

	delete(u.repositories, name)
	return nil
}

func (u *unitOfWork) Has(name RepositoryName) bool {
	_, ok := u.repositories[name]
	return ok
}

func (u *unitOfWork) Clear() {
	u.repositories = make(map[RepositoryName]RepositoryFactory)
}

func (u *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context, tx TX) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = fn(ctx, NewTransaction(tx, u.repositories))
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
