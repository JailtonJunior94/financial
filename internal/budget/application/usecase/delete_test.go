package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type DeleteBudgetUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	fm   *metrics.FinancialMetrics
	repo *repositoryMock.BudgetRepository
}

func TestDeleteBudgetUseCaseSuite(t *testing.T) {
	suite.Run(t, new(DeleteBudgetUseCaseSuite))
}

func (s *DeleteBudgetUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
}

func (s *DeleteBudgetUseCaseSuite) TestExecute() {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validBudgetID := "770e8400-e29b-41d4-a716-446655440002"
	infraErr := errors.New("database error")

	referenceMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	totalAmount, _ := vos.NewMoneyFromFloat(5000.00, vos.CurrencyBRL)
	userIDVO := mustParseUUID(validUserID)
	budgetIDVO := mustParseUUID(validBudgetID)

	type args struct {
		userID   string
		budgetID string
	}
	type dependencies func()

	scenarios := []struct {
		name         string
		uow          uow.UnitOfWork
		args         args
		dependencies dependencies
		expect       func(err error)
	}{
		{
			name: "should delete budget successfully",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				budget := buildTestBudget(userIDVO, totalAmount, referenceMonth)
				s.repo.EXPECT().
					FindByID(s.ctx, userIDVO, budgetIDVO).
					Return(budget, nil).
					Once()
				s.repo.EXPECT().
					Delete(s.ctx, budgetIDVO).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should return error when budget is not found",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				s.repo.EXPECT().
					FindByID(s.ctx, userIDVO, budgetIDVO).
					Return(nil, nil).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, domain.ErrBudgetNotFound))
			},
		},
		{
			name: "should return error when repository FindByID fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				s.repo.EXPECT().
					FindByID(s.ctx, userIDVO, budgetIDVO).
					Return(nil, infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when repository Delete fails",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: validBudgetID},
			dependencies: func() {
				budget := buildTestBudget(userIDVO, totalAmount, referenceMonth)
				s.repo.EXPECT().
					FindByID(s.ctx, userIDVO, budgetIDVO).
					Return(budget, nil).
					Once()
				s.repo.EXPECT().
					Delete(s.ctx, budgetIDVO).
					Return(infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when user_id is invalid",
			uow:  &passThroughUoW{},
			args: args{userID: "invalid-uuid", budgetID: validBudgetID},
			dependencies: func() {
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid user_id")
			},
		},
		{
			name: "should return error when budget_id is invalid",
			uow:  &passThroughUoW{},
			args: args{userID: validUserID, budgetID: "invalid-uuid"},
			dependencies: func() {
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid budget_id")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewDeleteBudgetUseCase(scenario.uow, s.obs, s.fm, s.repo)
			err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.budgetID)
			scenario.expect(err)
		})
	}
}
