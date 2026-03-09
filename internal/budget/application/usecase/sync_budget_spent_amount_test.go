package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/suite"

	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type SyncBudgetSpentAmountUseCaseSuite struct {
	suite.Suite
	ctx                  context.Context
	obs                  *fake.Provider
	fm                   *metrics.FinancialMetrics
	repo                 *repositoryMock.BudgetRepository
	invoiceCategoryTotal *repositoryMock.InvoiceCategoryTotalProvider
}

func TestSyncBudgetSpentAmountUseCaseSuite(t *testing.T) {
	suite.Run(t, new(SyncBudgetSpentAmountUseCaseSuite))
}

func (s *SyncBudgetSpentAmountUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewTestFinancialMetrics()
	s.repo = repositoryMock.NewBudgetRepository(s.T())
	s.invoiceCategoryTotal = repositoryMock.NewInvoiceCategoryTotalProvider(s.T())
}

func (s *SyncBudgetSpentAmountUseCaseSuite) TestExecute() {
	infraErr := errors.New("database error")
	invoiceErr := errors.New("invoice provider error")

	referenceMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	userIDVO := mustParseUUID("550e8400-e29b-41d4-a716-446655440000")
	categoryIDVO := mustParseUUID("660e8400-e29b-41d4-a716-446655440001")
	spentAmount, _ := vos.NewMoneyFromFloat(2000.00, vos.CurrencyBRL)

	type args struct {
		userID         vos.UUID
		referenceMonth pkgVos.ReferenceMonth
		categoryID     vos.UUID
	}
	type dependencies func()

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(err error)
	}{
		{
			name: "should sync spent amount successfully",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 0)
				budget.Items[0].CategoryID = categoryIDVO

				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(budget, nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(s.ctx, budget.Items[0]).
					Return(nil).
					Once()
				s.repo.EXPECT().
					Update(s.ctx, budget).
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should ignore silently when budget not found for user/month",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(nil, nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should ignore silently when category not linked to any budget item",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				otherCategoryID := mustParseUUID("770e8400-e29b-41d4-a716-446655440003")
				budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 0)
				budget.Items[0].CategoryID = otherCategoryID

				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(budget, nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should return error when invoice category total provider fails",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				zeroMoney, _ := vos.NewMoney(0, vos.CurrencyBRL)
				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(zeroMoney, invoiceErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, invoiceErr))
			},
		},
		{
			name: "should return error when repository FindByUserIDAndReferenceMonth fails",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(nil, infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when repository UpdateItem fails",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 0)
				budget.Items[0].CategoryID = categoryIDVO

				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(budget, nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(s.ctx, budget.Items[0]).
					Return(infraErr).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.True(errors.Is(err, infraErr))
			},
		},
		{
			name: "should return error when repository Update fails",
			args: args{
				userID:         userIDVO,
				referenceMonth: referenceMonth,
				categoryID:     categoryIDVO,
			},
			dependencies: func() {
				budget := buildBudgetWithItem(userIDVO, 5000.00, referenceMonth, 100_000, 0)
				budget.Items[0].CategoryID = categoryIDVO

				s.invoiceCategoryTotal.EXPECT().
					GetCategoryTotal(s.ctx, userIDVO, referenceMonth, categoryIDVO).
					Return(spentAmount, nil).
					Once()
				s.repo.EXPECT().
					FindByUserIDAndReferenceMonth(s.ctx, userIDVO, referenceMonth).
					Return(budget, nil).
					Once()
				s.repo.EXPECT().
					UpdateItem(s.ctx, budget.Items[0]).
					Return(nil).
					Once()
				s.repo.EXPECT().
					Update(s.ctx, budget).
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
			uc := NewSyncBudgetSpentAmountUseCase(
				&passThroughUoW{},
				s.invoiceCategoryTotal,
				s.repo,
				s.obs,
				s.fm,
			)
			err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.referenceMonth, scenario.args.categoryID)
			scenario.expect(err)
		})
	}
}
