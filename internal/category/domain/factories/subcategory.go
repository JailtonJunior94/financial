package factories

import (
	"fmt"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
)

func CreateSubcategory(userID, categoryID, name string, sequence uint) (*entities.Subcategory, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating subcategory id: %v", err)
	}

	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %s", userID)
	}

	catID, err := sharedVos.NewUUIDFromString(categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id: %s", categoryID)
	}

	subcategoryName, err := vos.NewCategoryName(name)
	if err != nil {
		return nil, err
	}

	sequenceVO, err := vos.NewCategorySequence(sequence)
	if err != nil {
		return nil, err
	}

	subcategory, err := entities.NewSubcategory(user, catID, subcategoryName, sequenceVO)
	if err != nil {
		return nil, fmt.Errorf("error creating subcategory: %w", err)
	}

	subcategory.ID = id
	return subcategory, nil
}
