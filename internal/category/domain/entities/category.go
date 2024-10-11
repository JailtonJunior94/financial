package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type Category struct {
	entity.Base
	UserID   sharedVos.UUID
	ParentID *sharedVos.UUID
	Name     string
	Sequence uint
}

func NewCategory(userID sharedVos.UUID, parentID *sharedVos.UUID, name string, sequence uint) (*Category, error) {
	category := &Category{
		UserID:   userID,
		ParentID: parentID,
		Name:     name,
		Sequence: sequence,
		Base: entity.Base{
			CreatedAt: time.Now(),
		},
	}
	return category, nil
}
