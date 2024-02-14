package repository

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/category/domain/entity"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
)

type categoryRepository struct {
	Db *sql.DB
}

func NewCategoryRepository(db *sql.DB) interfaces.CategoryRepository {
	return &categoryRepository{Db: db}
}

func (r *categoryRepository) Find(userID string) ([]*entity.Category, error) {
	return nil, nil
}

func (r *categoryRepository) FindByID(userID, id string) (*entity.Category, error) {
	return nil, nil
}

func (r *categoryRepository) Create(c *entity.Category) (*entity.Category, error) {
	stmt, err := r.Db.Prepare("insert into categories (id, user_id, name, sequence, created_at, updated_at, active) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(c.ID, c.UserID, c.Name, c.Sequence, c.CreatedAt, c.UpdatedAt, c.Active)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *categoryRepository) Update(c *entity.Category) (*entity.Category, error) {
	return nil, nil
}
