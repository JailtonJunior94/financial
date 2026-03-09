package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type ReplicateBudgetUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	repo *repositoryMock.BudgetRepository
}

func TestReplicateBudgetUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ReplicateBudgetUseCaseSuite))
}

func (s *ReplicateBudgetUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
}

func (s *ReplicateBudgetUseCaseSuite) TestExecute() {
	sourceBudget := s.buildSourceBudget("2026-03")
	existingBudget := s.buildSourceBudget("2026-04")
	expectedNextMonth, _ := pkgVos.NewReferenceMonth("2026-04")
	infraErr := errors.New("database error")

	type args struct {
		sourceBudget *entities.Budget
	}
	type dependencies func()
	type expect func(err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should replicate budget to next month when next month does not exist",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should skip replication when next month budget already exists",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(existingBudget, nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should replicate items with zero spent_amount and recalculated planned_amount",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Run(func(_ context.Context, b *entities.Budget) {
						for _, item := range b.Items {
							s.True(item.SpentAmount.IsZero(), "spent_amount must be zero for replicated items")
							s.False(item.PlannedAmount.IsZero(), "planned_amount must be recalculated")
						}
					}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should set next month as reference month in replicated budget",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Run(func(_ context.Context, b *entities.Budget) {
						s.True(b.ReferenceMonth.Equal(expectedNextMonth), "replicated budget must have next month as reference month")
					}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should generate new UUIDs for replicated budget and items",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Run(func(_ context.Context, b *entities.Budget) {
						s.NotEqual(sourceBudget.ID.String(), b.ID.String(), "replicated budget must have a new UUID")
						for i, item := range b.Items {
							s.NotEqual(sourceBudget.Items[i].ID.String(), item.ID.String(), "replicated item must have a new UUID")
						}
					}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should return error when finding next month budget fails",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when inserting replicated budget fails",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when inserting replicated items fails",
			args: args{sourceBudget: sourceBudget},
			dependencies: func() {
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, sourceBudget.UserID, expectedNextMonth).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
				s.repo.EXPECT().
					InsertItems(mock.Anything, mock.AnythingOfType("[]*entities.BudgetItem")).
					Return(infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewReplicateBudgetUseCase(s.obs)
			err := uc.Execute(s.ctx, s.repo, scenario.args.sourceBudget)
			scenario.expect(err)
		})
	}
}

func (s *ReplicateBudgetUseCaseSuite) buildSourceBudget(referenceMonthStr string) *entities.Budget {
	userID, err := vos.NewUUID()
	s.Require().NoError(err)

	budgetID, err := vos.NewUUID()
	s.Require().NoError(err)

	totalAmount, err := vos.NewMoneyFromFloat(10000.00, vos.CurrencyBRL)
	s.Require().NoError(err)

	referenceMonth, err := pkgVos.NewReferenceMonth(referenceMonthStr)
	s.Require().NoError(err)

	budget := entities.NewBudget(userID, totalAmount, referenceMonth)
	budget.SetID(budgetID)

	category1ID, err := vos.NewUUID()
	s.Require().NoError(err)

	category2ID, err := vos.NewUUID()
	s.Require().NoError(err)

	item1ID, err := vos.NewUUID()
	s.Require().NoError(err)

	item2ID, err := vos.NewUUID()
	s.Require().NoError(err)

	percentage40, err := vos.NewPercentage(40000)
	s.Require().NoError(err)

	percentage60, err := vos.NewPercentage(60000)
	s.Require().NoError(err)

	item1 := entities.NewBudgetItem(budget.ID, totalAmount, category1ID, percentage40)
	item1.SetID(item1ID)

	item2 := entities.NewBudgetItem(budget.ID, totalAmount, category2ID, percentage60)
	item2.SetID(item2ID)

	err = budget.AddItems([]*entities.BudgetItem{item1, item2})
	s.Require().NoError(err)

	return budget
}
