package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type FindBudgetUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	fm   *metrics.FinancialMetrics
	repo *repositoryMock.BudgetRepository
}

func TestFindBudgetUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindBudgetUseCaseSuite))
}

func (s *FindBudgetUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
}

func buildBudgetWithItem(userID vos.UUID, totalAmountFloat float64, referenceMonth pkgVos.ReferenceMonth, percentageGoalInt64 int64, spentAmountFloat float64) *entities.Budget {
	totalAmount, _ := vos.NewMoneyFromFloat(totalAmountFloat, vos.CurrencyBRL)
	budget := buildTestBudget(userID, totalAmount, referenceMonth)

	categoryID, _ := vos.NewUUID()
	percentageGoal, _ := vos.NewPercentage(percentageGoalInt64)
	item := entities.NewBudgetItem(budget.ID, totalAmount, categoryID, percentageGoal)
	item.SetID(mustNewUUID())

	if spentAmountFloat > 0 {
		spent, _ := vos.NewMoneyFromFloat(spentAmountFloat, vos.CurrencyBRL)
		item.SpentAmount = spent
	}

	budget.Items = append(budget.Items, item)
	_ = budget.RecalculateTotals()
	return budget
}

func mustNewUUID() vos.UUID {
	id, _ := vos.NewUUID()
	return id
}

func (s *FindBudgetUseCaseSuite) TestExecute() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validBudgetID := "770e8400-e29b-41d4-a716-446655440002"
	infraErr := errors.New("database error")

	referenceMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	userIDVO, _ := vos.NewUUIDFromString(validUserID)

	type args struct {
		userID   string
		budgetID string
	}
	type dependencies func()

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output any, err error)
	}{
		{
			name: "should return budget with calculated fields correctly",
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 2000.00)
				s.repo.EXPECT().
					FindByID(
						s.ctx,
						userIDVO,
						mustParseUUID(validBudgetID),
					).
					Return(budget, nil).
					Once()
			},
			expect: func(output any, err error) {
				s.NoError(err)
				s.NotNil(output)
			},
		},
		{
			name: "should return error when budget is not found",
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				s.repo.EXPECT().
					FindByID(
						s.ctx,
						userIDVO,
						mustParseUUID(validBudgetID),
					).
					Return(nil, nil).
					Once()
			},
			expect: func(output any, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrBudgetNotFound))
			},
		},
		{
			name: "should return error when repository fails",
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				s.repo.EXPECT().
					FindByID(
						s.ctx,
						userIDVO,
						mustParseUUID(validBudgetID),
					).
					Return(nil, infraErr).
					Once()
			},
			expect: func(output any, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when user_id is invalid",
			args: args{userID: "invalid-uuid", budgetID: validBudgetID},
			dependencies: func() {
			},
			expect: func(output any, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid user_id")
			},
		},
		{
			name: "should return error when budget_id is invalid",
			args: args{userID: validUserID, budgetID: "invalid-uuid"},
			dependencies: func() {
			},
			expect: func(output any, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid budget_id")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewFindBudgetUseCase(s.repo, s.obs, s.fm)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.budgetID)
			scenario.expect(output, err)
		})
	}
}

func (s *FindBudgetUseCaseSuite) TestCalculatedFields() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validBudgetID := "770e8400-e29b-41d4-a716-446655440002"

	referenceMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	userIDVO, _ := vos.NewUUIDFromString(validUserID)

	s.Run("item with no spent amount shows remaining equal to planned and percentage_spent zero", func() {
		budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 0)
		s.repo.EXPECT().
			FindByID(s.ctx, userIDVO, mustParseUUID(validBudgetID)).
			Return(budget, nil).
			Once()

		uc := NewFindBudgetUseCase(s.repo, s.obs, s.fm)
		output, err := uc.Execute(s.ctx, validUserID, validBudgetID)

		s.NoError(err)
		s.Require().NotNil(output)
		s.Require().Len(output.Items, 1)

		item := output.Items[0]
		s.Equal(item.PlannedAmount, item.RemainingAmount, "remaining should equal planned when no spent")
		s.Equal("0.000", item.PercentageSpent, "percentage_spent should be 0 when nothing is spent")
	})

	s.Run("item with spent above planned shows percentage_spent above 100 and negative remaining", func() {
		budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 6000.00)
		s.repo.EXPECT().
			FindByID(s.ctx, userIDVO, mustParseUUID(validBudgetID)).
			Return(budget, nil).
			Once()

		uc := NewFindBudgetUseCase(s.repo, s.obs, s.fm)
		output, err := uc.Execute(s.ctx, validUserID, validBudgetID)

		s.NoError(err)
		s.Require().NotNil(output)
		s.Require().Len(output.Items, 1)

		item := output.Items[0]
		percentageSpent, parseErr := strconv.ParseFloat(item.PercentageSpent, 64)
		s.NoError(parseErr)
		s.Greater(percentageSpent, 100.0, "percentage_spent should be above 100 when spent exceeds planned")
		s.True(strings.HasPrefix(item.RemainingAmount, "-"), "remaining_amount should be negative when spent exceeds planned")
	})

	s.Run("item with partial spending shows correct remaining and percentage", func() {
		budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 2500.00)
		s.repo.EXPECT().
			FindByID(s.ctx, userIDVO, mustParseUUID(validBudgetID)).
			Return(budget, nil).
			Once()

		uc := NewFindBudgetUseCase(s.repo, s.obs, s.fm)
		output, err := uc.Execute(s.ctx, validUserID, validBudgetID)

		s.NoError(err)
		s.Require().NotNil(output)
		s.Require().Len(output.Items, 1)

		item := output.Items[0]
		s.Equal("2500.00", item.RemainingAmount)
		s.Equal("50.000", item.PercentageSpent)
	})
}

func mustParseUUID(id string) vos.UUID {
	uid, _ := vos.NewUUIDFromString(id)
	return uid
}
