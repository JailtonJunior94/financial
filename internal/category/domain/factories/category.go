package factories

import (
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

func CreateCategory(userID, parentID, name string, sequence uint) (*entities.Category, error) {
	id, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	var p *vos.UUID
	if len(parentID) > 0 {
		parent, err := vos.NewUUIDFromString(parentID)
		if err != nil {
			return nil, err
		}
		p = &parent
	}

	category, err := entities.NewCategory(user, p, name, sequence)
	if err != nil {
		return nil, err
	}

	category.ID = id
	return category, nil
}
