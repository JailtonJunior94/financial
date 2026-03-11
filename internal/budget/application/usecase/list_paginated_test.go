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
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type ListBudgetsPaginatedUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	fm   *metrics.FinancialMetrics
	repo *repositoryMock.BudgetRepository
}

func TestListBudgetsPaginatedUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ListBudgetsPaginatedUseCaseSuite))
}

func (s *ListBudgetsPaginatedUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
}

func (s *ListBudgetsPaginatedUseCaseSuite) buildBudgets(userID vos.UUID, count int) []*entities.Budget {
	budgets := make([]*entities.Budget, count)
	months := []string{"2026-01", "2026-02", "2026-03", "2026-04", "2026-05"}
	totalAmount, _ := vos.NewMoneyFromFloat(5000.00, vos.CurrencyBRL)
	for i := 0; i < count; i++ {
		month, _ := pkgVos.NewReferenceMonth(months[i%len(months)])
		budgets[i] = buildTestBudget(userID, totalAmount, month)
	}
	return budgets
}

func (s *ListBudgetsPaginatedUseCaseSuite) TestExecute() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	infraErr := errors.New("database error")

	userIDVO := mustParseUUID(validUserID)

	type args struct {
		input ListBudgetsPaginatedInput
	}
	type dependencies func()

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *ListBudgetsPaginatedOutput, err error)
	}{
		{
			name: "should list budgets with pagination",
			args: args{
				input: ListBudgetsPaginatedInput{
					UserID: validUserID,
					Limit:  2,
					Cursor: "",
				},
			},
			dependencies: func() {
				budgets := s.buildBudgets(userIDVO, 2)
				s.repo.EXPECT().
					ListPaginated(s.ctx, mock.MatchedBy(func(p interfaces.ListBudgetsParams) bool {
						return p.UserID == userIDVO && p.Limit == 3
					})).
					Return(budgets, nil).
					Once()
			},
			expect: func(output *ListBudgetsPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Budgets, 2)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "should return next cursor when more results exist",
			args: args{
				input: ListBudgetsPaginatedInput{
					UserID: validUserID,
					Limit:  2,
					Cursor: "",
				},
			},
			dependencies: func() {
				budgets := s.buildBudgets(userIDVO, 3)
				s.repo.EXPECT().
					ListPaginated(s.ctx, mock.MatchedBy(func(p interfaces.ListBudgetsParams) bool {
						return p.UserID == userIDVO && p.Limit == 3
					})).
					Return(budgets, nil).
					Once()
			},
			expect: func(output *ListBudgetsPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Budgets, 2)
				s.NotNil(output.NextCursor)
			},
		},
		{
			name: "should return empty list when no budgets exist",
			args: args{
				input: ListBudgetsPaginatedInput{
					UserID: validUserID,
					Limit:  10,
					Cursor: "",
				},
			},
			dependencies: func() {
				s.repo.EXPECT().
					ListPaginated(s.ctx, mock.MatchedBy(func(p interfaces.ListBudgetsParams) bool {
						return p.UserID == userIDVO && p.Limit == 11
					})).
					Return([]*entities.Budget{}, nil).
					Once()
			},
			expect: func(output *ListBudgetsPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Empty(output.Budgets)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "should return error when repository fails",
			args: args{
				input: ListBudgetsPaginatedInput{
					UserID: validUserID,
					Limit:  10,
					Cursor: "",
				},
			},
			dependencies: func() {
				s.repo.EXPECT().
					ListPaginated(s.ctx, mock.MatchedBy(func(p interfaces.ListBudgetsParams) bool {
						return p.UserID == userIDVO && p.Limit == 11
					})).
					Return(nil, infraErr).
					Once()
			},
			expect: func(output *ListBudgetsPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when user_id is invalid",
			args: args{
				input: ListBudgetsPaginatedInput{
					UserID: "invalid-uuid",
					Limit:  10,
					Cursor: "",
				},
			},
			dependencies: func() {
			},
			expect: func(output *ListBudgetsPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewListBudgetsPaginatedUseCase(s.obs, s.fm, s.repo)
			output, err := uc.Execute(s.ctx, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
