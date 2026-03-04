package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/stretchr/testify/suite"
)

type DeleteUserUseCaseSuite struct {
	suite.Suite
	ctx            context.Context
	userRepository *repositoryMock.UserRepository
}

func TestDeleteUserUseCaseSuite(t *testing.T) {
	suite.Run(t, new(DeleteUserUseCaseSuite))
}

func (s *DeleteUserUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepository = repositoryMock.NewUserRepository(s.T())
}

func (s *DeleteUserUseCaseSuite) TestExecute() {
	type args struct {
		id string
	}

	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(err error)
	}{
		{
			name: "deve deletar usuário com sucesso",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					SoftDelete(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil).
					Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "deve retornar ErrUserNotFound quando usuário não existe",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					SoftDelete(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(customerrors.ErrUserNotFound).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, customerrors.ErrUserNotFound)
			},
		},
		{
			name: "deve retornar erro de infra ao falhar no repositório",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					SoftDelete(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(errors.New("db error")).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "db error")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.setup()

			obs := fake.NewProvider()
			fm := metrics.NewTestFinancialMetrics()
			uc := NewDeleteUserUseCase(obs, fm, s.userRepository)
			err := uc.Execute(s.ctx, scenario.args.id)

			scenario.expect(err)
		})
	}
}
