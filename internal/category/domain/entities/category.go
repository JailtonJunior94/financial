package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"

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

// Update atualiza os dados da categoria.
// parentID pode ser nil para categorias raiz (sem pai).
// Se parentID for fornecido, não pode ser um UUID vazio (uuid.Nil).
func (c *Category) Update(name string, sequence uint, parentID *sharedVos.UUID) error {
	categoryName, err := vos.NewCategoryName(name)
	if err != nil {
		return err
	}

	categorySequence, err := vos.NewCategorySequence(sequence)
	if err != nil {
		return err
	}

	// Validar parentID: nil é válido (categoria raiz), mas se fornecido não pode ser UUID vazio
	if parentID != nil && parentID.Value == uuid.Nil {
		return errors.New("parent ID cannot be empty UUID")
	}

	c.Name = categoryName
	c.Sequence = categorySequence
	c.ParentID = parentID
	c.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

func (c *Category) Delete() *Category {
	c.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return c
}
