package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	userdomain "github.com/jailtonjunior94/financial/internal/user/domain"
	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetUserUseCaseSuite struct {
	suite.Suite
	ctx            context.Context
	userRepository *repositoryMock.UserRepository
}

func TestGetUserUseCaseSuite(t *testing.T) {
	suite.Run(t, new(GetUserUseCaseSuite))
}

func (s *GetUserUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepository = repositoryMock.NewUserRepository(s.T())
}

func newTestUserEntity(t *testing.T) *entities.User {
	t.Helper()
	name, err := vos.NewUserName("John Doe")
	require.NoError(t, err)
	email, err := vos.NewEmail("john@example.com")
	require.NoError(t, err)
	user, err := entities.NewUser(name, email)
	require.NoError(t, err)
	require.NoError(t, user.SetPassword("hashed_password_123456789"))
	return user
}

func (s *GetUserUseCaseSuite) TestExecute() {
	type args struct {
		id string
	}

	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(output interface{}, err error)
	}{
		{
			name: "deve retornar usuário com sucesso",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.NoError(err)
				s.NotNil(output)
			},
		},
		{
			name: "deve retornar ErrUserNotFound quando usuário não existe",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, userdomain.ErrUserNotFound)
			},
		},
		{
			name: "deve retornar erro de infra ao falhar no repositório",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, errors.New("db error")).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "db error")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.setup()

			obs := fake.NewProvider()
			fm := metrics.NewTestFinancialMetrics()
			uc := NewGetUserUseCase(obs, fm, s.userRepository)
			output, err := uc.Execute(s.ctx, scenario.args.id)

			scenario.expect(output, err)
		})
	}
}
