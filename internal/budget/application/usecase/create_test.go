package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
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

func buildTestBudget(userID vos.UUID, totalAmount vos.Money, referenceMonth pkgVos.ReferenceMonth) *entities.Budget {
	budgetID, _ := vos.NewUUID()
	budget := entities.NewBudget(userID, totalAmount, referenceMonth)
	budget.SetID(budgetID)
	return budget
}

type passThroughUoW struct{}

func (m *passThroughUoW) Do(ctx context.Context, fn func(ctx context.Context, tx database.DBTX) error) error {
	return fn(ctx, nil)
}

type errorUoW struct {
	err error
}

func (m *errorUoW) Do(_ context.Context, _ func(ctx context.Context, tx database.DBTX) error) error {
	return m.err
}

var _ uow.UnitOfWork = (*passThroughUoW)(nil)
var _ uow.UnitOfWork = (*errorUoW)(nil)

type CreateBudgetUseCaseSuite struct {
	suite.Suite
	ctx              context.Context
	obs              *fake.Provider
	fm               *metrics.FinancialMetrics
	repo             *repositoryMock.BudgetRepository
	categoryProvider *repositoryMock.CategoryProvider
	replicateUC      *repositoryMock.ReplicateBudgetUseCase
}

func TestCreateBudgetUseCaseSuite(t *testing.T) {
	suite.Run(t, new(CreateBudgetUseCaseSuite))
}

func (s *CreateBudgetUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
	s.categoryProvider = repositoryMock.NewCategoryProvider(s.T())
	s.replicateUC = repositoryMock.NewReplicateBudgetUseCase(s.T())
}

func (s *CreateBudgetUseCaseSuite) TestExecute() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validCategoryID := "660e8400-e29b-41d4-a716-446655440001"
	infraErr := errors.New("database error")
	replicateErr := errors.New("replicate error")

	validInput := func() *dtos.BudgetCreateInput {
		return &dtos.BudgetCreateInput{
			ReferenceMonth: "2026-03",
			TotalAmount:    "5000.00",
			Currency:       "BRL",
			Items: []dtos.BudgetItemInput{
				{CategoryID: validCategoryID, PercentageGoal: "100.000"},
			},
		}
	}

	type args struct {
		userID string
		input  *dtos.BudgetCreateInput
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
			name: "should create budget successfully",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.ReferenceMonth")).
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
				s.replicateUC.EXPECT().
					Execute(mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Budget")).
					Return(nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(validUserID, output.UserID)
				s.Equal("2026-03", output.ReferenceMonth)
				s.Equal("5000.00", output.TotalAmount)
				s.Len(output.Items, 1)
			},
		},
		{
			name: "should return error when category validation fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
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
			name: "should return error when factory fails due to invalid currency",
			uow:  &passThroughUoW{},
			args: args{
				userID: validUserID,
				input: &dtos.BudgetCreateInput{
					ReferenceMonth: "2026-03",
					TotalAmount:    "5000.00",
					Currency:       "INVALID",
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
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "unsupported currency")
			},
		},
		{
			name: "should return error when budget already exists for reference month",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()

				existingMonth, _ := pkgVos.NewReferenceMonth("2026-03")
				existingUserID, _ := vos.NewUUIDFromString(validUserID)
				existingTotalAmount, _ := vos.NewMoneyFromFloat(5000, vos.CurrencyBRL)
				existing := buildTestBudget(existingUserID, existingTotalAmount, existingMonth)

				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.ReferenceMonth")).
					Return(existing, nil).
					Once()
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrBudgetAlreadyExistsForMonth))
			},
		},
		{
			name: "should return error when repository Insert fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.ReferenceMonth")).
					Return(nil, nil).
					Once()
				s.repo.EXPECT().
					Insert(mock.Anything, mock.AnythingOfType("*entities.Budget")).
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
			name: "should return error when repository InsertItems fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.ReferenceMonth")).
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
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error and rollback when replication fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, input: validInput()},
			dependencies: func() {
				s.categoryProvider.EXPECT().
					ValidateCategories(mock.Anything, validUserID, []string{validCategoryID}).
					Return(nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(mock.Anything, mock.AnythingOfType("vos.UUID"), mock.AnythingOfType("vos.ReferenceMonth")).
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
			name: "should return error when percentages do not sum to 100",
			uow:  &passThroughUoW{},
			args: args{
				userID: validUserID,
				input: &dtos.BudgetCreateInput{
					ReferenceMonth: "2026-03",
					TotalAmount:    "5000.00",
					Currency:       "BRL",
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
			},
			expect: func(output *dtos.BudgetOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.True(errors.Is(err, domain.ErrBudgetInvalidTotal))
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewCreateBudgetUseCase(
				scenario.uow,
				s.obs,
				s.fm,
				s.repo,
				s.categoryProvider,
				s.replicateUC,
			)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
