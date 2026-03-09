package usecase

import (
	"context"
	"errors"
	"testing"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

func buildBudgetWithItems(userID vos.UUID, totalAmount vos.Money, referenceMonth pkgVos.ReferenceMonth, items []*entities.BudgetItem) *entities.Budget {
	budget := buildTestBudget(userID, totalAmount, referenceMonth)
	budget.Items = items
	return budget
}

func buildBudgetItem(budgetID vos.UUID, totalAmount vos.Money, categoryID vos.UUID, percentageRaw int64, spentAmountFloat float64) *entities.BudgetItem {
	itemID, _ := vos.NewUUID()
	percentage, _ := vos.NewPercentage(percentageRaw)
	item := entities.NewBudgetItem(budgetID, totalAmount, categoryID, percentage)
	item.SetID(itemID)
	if spentAmountFloat > 0 {
		spentAmount, _ := vos.NewMoneyFromFloat(spentAmountFloat, totalAmount.Currency())
		item.SpentAmount = spentAmount
	}
	return item
}

type UpdateBudgetUseCaseSuite struct {
	suite.Suite
	ctx              context.Context
	obs              *fake.Provider
	fm               *metrics.FinancialMetrics
	repo             *repositoryMock.BudgetRepository
	categoryProvider *repositoryMock.CategoryProvider
	replicateUC      *repositoryMock.ReplicateBudgetUseCase
}

func TestUpdateBudgetUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateBudgetUseCaseSuite))
}

func (s *UpdateBudgetUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
	s.categoryProvider = repositoryMock.NewCategoryProvider(s.T())
	s.replicateUC = repositoryMock.NewReplicateBudgetUseCase(s.T())
}

func (s *UpdateBudgetUseCaseSuite) TestExecute() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validBudgetID := "660e8400-e29b-41d4-a716-446655440001"
	validCategoryID := "770e8400-e29b-41d4-a716-446655440002"
	newCategoryID := "880e8400-e29b-41d4-a716-446655440003"
	infraErr := errors.New("database error")
	replicateErr := errors.New("replicate error")

	parsedUserID, _ := vos.NewUUIDFromString(validUserID)
	parsedCategoryID, _ := vos.NewUUIDFromString(validCategoryID)
	totalAmount, _ := vos.NewMoneyFromFloat(5000, vos.CurrencyBRL)
	referenceMonth, _ := pkgVos.NewReferenceMonth("2026-03")

	validInput := func() *dtos.BudgetUpdateInput {
		return &dtos.BudgetUpdateInput{
			TotalAmount: "5000.00",
			Items: []dtos.BudgetItemInput{
				{CategoryID: validCategoryID, PercentageGoal: "100.000"},
			},
		}
	}

	buildExistingBudget := func(spentAmount float64) *entities.Budget {
		item := buildBudgetItem(parsedUserID, totalAmount, parsedCategoryID, 100_000, spentAmount)
		return buildBudgetWithItems(parsedUserID, totalAmount, referenceMonth, []*entities.BudgetItem{item})
	}

	type args struct {
		userID   string
		budgetID string
		input    *dtos.BudgetUpdateInput
	}
	type dependencies func()
	type expect func(output *dtos.BudgetOutput, err error)

	scenarios := []struct {
		name         string
		uow          uow.UnitOfWork
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should update budget successfully recalculating planned_amount",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(1000.00)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(validUserID, output.UserID)
				s.Equal("5000.00", output.TotalAmount)
				s.Len(output.Items, 1)
				s.Equal("5000.00", output.Items[0].PlannedAmount)
				s.Equal("1000.00", output.Items[0].SpentAmount)
			},
		},
		{
			name: "should update total_amount and recalculate planned_amount preserving spent_amount",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "10000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: validCategoryID, PercentageGoal: "100.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(1000.00)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("10000.00", output.TotalAmount)
				s.Equal("10000.00", output.Items[0].PlannedAmount)
				s.Equal("1000.00", output.Items[0].SpentAmount)
			},
		},
		{
			name: "should redistribute percentages and recalculate planned_amount",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "5000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: validCategoryID, PercentageGoal: "60.000"},
						{CategoryID: newCategoryID, PercentageGoal: "40.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID, newCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(500.00)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Items, 2)
				s.Equal("5000.00", output.TotalAmount)
			},
		},
		{
			name: "should preserve spent_amount for existing items and zero for new items",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "5000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: validCategoryID, PercentageGoal: "50.000"},
						{CategoryID: newCategoryID, PercentageGoal: "50.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID, newCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(800.00)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Items, 2)
				var existingItemOutput, newItemOutput *dtos.BudgetItemOutput
				for i := range output.Items {
					if output.Items[i].CategoryID == validCategoryID {
						existingItemOutput = &output.Items[i]
					} else {
						newItemOutput = &output.Items[i]
					}
				}
				s.Require().NotNil(existingItemOutput)
				s.Require().NotNil(newItemOutput)
				s.Equal("800.00", existingItemOutput.SpentAmount)
				s.Equal("0.00", newItemOutput.SpentAmount)
			},
		},
		{
			name: "should return error when percentages do not sum to 100",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "5000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: validCategoryID, PercentageGoal: "50.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(0)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrBudgetInvalidTotal))
			},
		},
		{
			name: "should return error when category validation fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(domain.ErrCategoryNotFound).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrCategoryNotFound))
			},
		},
		{
			name: "should return error when budget is not found",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(nil, nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrBudgetNotFound))
			},
		},
		{
			name: "should return error when category_id is duplicated in input",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "5000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: validCategoryID, PercentageGoal: "50.000"},
						{CategoryID: validCategoryID, PercentageGoal: "50.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID, validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(0)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrDuplicateCategory))
			},
		},
		{
			name: "should return error when replication fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(0)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(replicateErr).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, replicateErr))
			},
		},
		{
			name: "should return error when repository FindByID fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(nil, infraErr).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when repository Update fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(0)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(infraErr).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name:         "should return error when invalid user_id",
			uow:          &passThroughUoW{},
			args:         args{userID: "not-a-uuid", budgetID: validBudgetID, input: validInput()},
			dependencies: func() {},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid user_id")
			},
		},
		{
			name: "should remove items not in update input via DeleteItemsNotIn",
			uow:  &passThroughUoW{},
			args: args{
				userID:   validUserID,
				budgetID: validBudgetID,
				input: &dtos.BudgetUpdateInput{
					TotalAmount: "5000.00",
					Items: []dtos.BudgetItemInput{
						{CategoryID: newCategoryID, PercentageGoal: "100.000"},
					},
				},
			},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{newCategoryID}).
					Return(nil).
					Once()
				// Budget has validCategoryID item; update replaces with newCategoryID only
				existingBudget := buildExistingBudget(500.00)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(nil).
					Once()
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Items, 1)
				s.Equal(newCategoryID, output.Items[0].CategoryID)
			},
		},
		{
			name: "should return error when DeleteItemsNotIn fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				existingBudget := buildExistingBudget(0)
				s.repo.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.UUID")).
					Return(existingBudget, nil).
					Once()
				s.repo.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(mock.Anything, mock.AnythingOfType("*entities.BudgetItem")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					DeleteItemsNotIn(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.Anything).
					Return(infraErr).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewUpdateBudgetUseCase(
				scenario.uow,
				s.obs,
				s.fm,
				s.repo,
				s.categoryProvider,
				s.replicateUC,
			)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.budgetID, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}

