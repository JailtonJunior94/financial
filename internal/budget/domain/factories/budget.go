package factories

import (
	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

func CreateBudget(userID string, input *dtos.BugetInput) (*entities.Budget, error) {
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	budgetID, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}

	budget := entities.NewBudget(user, vos.NewMoney(input.AmountGoal), input.Date)
	budget.SetID(budgetID)

	for _, item := range input.Items {
		category, err := vos.NewUUIDFromString(item.CategoryID)
		if err != nil {
			return nil, err
		}

		budgetItemID, err := vos.NewUUID()
		if err != nil {
			return nil, err
		}
		newItem := entities.NewBudgetItem(budget, category, vos.NewPercentage(item.PercentageGoal))
		newItem.SetID(budgetItemID)
		budget.AddItem(newItem)
	}
	return budget, nil
}
