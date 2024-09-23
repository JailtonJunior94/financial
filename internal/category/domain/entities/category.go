package entities

import (
	"errors"
	"time"

	"github.com/jailtonjunior94/financial/internal/shared/entity"
	sharedVos "github.com/jailtonjunior94/financial/internal/shared/vos"
)

var (
	ErrNameIsRequired     = errors.New("name is required")
	ErrUserIDIsRequired   = errors.New("user_id is required")
	ErrSequenceIsRequired = errors.New("sequence is required")
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
