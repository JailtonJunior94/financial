package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	userdomain "github.com/jailtonjunior94/financial/internal/user/domain"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
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
			name: "should delete user successfully",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
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
			name: "should return ErrUserNotFound when user does not exist",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, nil).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, userdomain.ErrUserNotFound)
			},
		},
		{
			name: "should return infra error when FindByID fails",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, errors.New("db error")).
					Once()
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "db error")
			},
		},
		{
			name: "should return infra error when SoftDelete fails",
			args: args{id: "550e8400-e29b-41d4-a716-446655440000"},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
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
