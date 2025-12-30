package factories

import (
	"fmt"
	"math"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

func CreateBudget(userID string, input *dtos.BugetInput) (*entities.Budget, error) {
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("create_budget: %v", err)
	}

	budgetID, err := vos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create_budget: %v", err)
	}

	// Convert amount from float to cents (int64)
	amountCents := int64(input.AmountGoal * 100)
	amountGoal, err := vos.NewMoney(amountCents, vos.CurrencyBRL) // Using BRL as default currency
	if err != nil {
		return nil, fmt.Errorf("create_budget: %v", err)
	}

	budget := entities.NewBudget(user, amountGoal, input.Date)
	budget.SetID(budgetID)

	// Validate that the sum of percentage_goal equals 100%
	var totalPercentage float64
	for _, item := range input.Items {
		totalPercentage += item.PercentageGoal
	}

	// Use a small epsilon for floating-point comparison
	const epsilon = 0.01
	if math.Abs(totalPercentage-100.0) > epsilon {
		return nil, customErrors.ErrBudgetInvalidTotal
	}

	for _, item := range input.Items {
		category, err := vos.NewUUIDFromString(item.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("create_budget: %v", err)
		}

		budgetItemID, err := vos.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("create_budget: %v", err)
		}

		// Convert percentage from float to int64 with scale 3 (12.5% → 12500)
		percentageInt := int64(item.PercentageGoal * 1000)
		percentage, err := vos.NewPercentage(percentageInt)
		if err != nil {
			return nil, fmt.Errorf("create_budget: %v", err)
		}

		newItem := entities.NewBudgetItem(budget, category, percentage)
		newItem.SetID(budgetItemID)

		// AddItem returns bool indicating if total percentage equals 100%
		// We already validated this above, so we can safely add items
		budget.AddItem(newItem)
	}
	return budget, nil
}
