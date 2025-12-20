package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
)

type Category struct {
	ID        sharedVos.UUID
	UserID    sharedVos.UUID
	ParentID  *sharedVos.UUID
	Name      vos.CategoryName
	Sequence  vos.CategorySequence
	Children  []Category
	CreatedAt sharedVos.NullableTime
	UpdatedAt sharedVos.NullableTime
	DeletedAt sharedVos.NullableTime
}

func NewCategory(userID sharedVos.UUID, parentID *sharedVos.UUID, name vos.CategoryName, sequence vos.CategorySequence) (*Category, error) {
	category := &Category{
		Name:      name,
		UserID:    userID,
		ParentID:  parentID,
		Sequence:  sequence,
		CreatedAt: sharedVos.NewNullableTime(time.Now()),
	}
	return category, nil
}

func (c *Category) AddChildrens(childrens []Category) {
	c.Children = childrens
}

func (c *Category) Update(name string, sequence uint, parentID *sharedVos.UUID) *Category {
	c.Name = vos.NewCategoryName(name)
	c.Sequence = vos.NewCategorySequence(sequence)
	c.ParentID = parentID
	c.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return c
}

func (c *Category) Delete() *Category {
	c.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return c
}
