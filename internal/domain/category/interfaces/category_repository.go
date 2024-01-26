package interfaces

import "github.com/jailtonjunior94/financial/internal/domain/category/entity"

type CategoryRepository interface {
	Find(userID string) ([]*entity.Category, error)
	FindByID(userID, id string) (*entity.Category, error)
	Create(c *entity.Category) (*entity.Category, error)
	Update(c *entity.Category) (*entity.Category, error)
}
