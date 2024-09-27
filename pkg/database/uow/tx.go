package uow

import "database/sql"

type transaction struct {
	tx           *sql.Tx
	repositories map[RepositoryName]RepositoryFactory
}

func NewTransaction(tx *sql.Tx, repositories map[RepositoryName]RepositoryFactory) *transaction {
	return &transaction{
		tx:           tx,
		repositories: repositories,
	}
}

func (t *transaction) Get(name RepositoryName) (Repository, error) {
	if repository, ok := t.repositories[name]; ok {
		return repository(t.tx), nil
	}
	return nil, ErrRepositoryNotRegistered
}
