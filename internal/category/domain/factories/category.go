package factories

import (
	"fmt"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
)

func CreateCategory(userID, parentID, name string, sequence uint) (*entities.Category, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating category id: %v", err)
	}

	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %s", userID)
	}

	var p *sharedVos.UUID
	if len(parentID) > 0 {
		parent, err := sharedVos.NewUUIDFromString(parentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %s", parentID)
		}
		p = &parent
	}

	categoryName, err := vos.NewCategoryName(name)
	if err != nil {
		return nil, err
	}

	sequenceVO, err := vos.NewCategorySequence(sequence)
	if err != nil {
		return nil, err
	}

	category, err := entities.NewCategory(user, p, categoryName, sequenceVO)
	if err != nil {
		return nil, fmt.Errorf("error creating category: %w", err)
	}

	category.ID = id
	return category, nil
}
