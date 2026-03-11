package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	categorydomain "github.com/jailtonjunior94/financial/internal/category/domain"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

// mockUnitOfWork is a simple inline mock for uow.UnitOfWork.
type mockUnitOfWork struct {
	returnErr error
}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, tx database.DBTX) error) error {
	return m.returnErr
}

type RemoveCategoryUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	fm                 *metrics.FinancialMetrics
	categoryRepository *mocks.CategoryRepository
}

func TestRemoveCategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(RemoveCategoryUseCaseSuite))
}

func (s *RemoveCategoryUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.fm = metrics.NewTestFinancialMetrics()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *RemoveCategoryUseCaseSuite) TestExecute() {
	type args struct {
		userID     string
		categoryID string
	}

	type dependencies struct {
		categoryRepository *mocks.CategoryRepository
		uow                *mockUnitOfWork
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(err error)
	}{
		{
			name: "deve remover categoria com sucesso",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
					return s.categoryRepository
				}(),
				uow: &mockUnitOfWork{returnErr: nil},
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "deve retornar erro quando categoria não for encontrada",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440099",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440099")
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(nil, nil).Once()
					return s.categoryRepository
				}(),
				uow: &mockUnitOfWork{},
			},
			expect: func(err error) {
				s.Error(err)
				s.Equal(categorydomain.ErrCategoryNotFound, err)
			},
		},
		{
			name: "deve retornar erro com user_id inválido",
			args: args{
				userID:     "invalid-uuid",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
				uow:                &mockUnitOfWork{},
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
		{
			name: "deve retornar erro com category_id inválido",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "invalid-uuid",
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
				uow:                &mockUnitOfWork{},
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			catFactory := func(tx database.DBTX) interfaces.CategoryRepository { return scenario.dependencies.categoryRepository }
			subcatFactory := func(tx database.DBTX) interfaces.SubcategoryRepository { return nil }
			uc := NewRemoveCategoryUseCase(s.obs, s.fm, scenario.dependencies.uow, scenario.dependencies.categoryRepository, catFactory, subcatFactory)
			err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID)
			scenario.expect(err)
		})
	}
}
