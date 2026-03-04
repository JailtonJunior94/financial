package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/stretchr/testify/suite"
)

type ListUsersUseCaseSuite struct {
	suite.Suite
	ctx            context.Context
	userRepository *repositoryMock.UserRepository
}

func TestListUsersUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ListUsersUseCaseSuite))
}

func (s *ListUsersUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepository = repositoryMock.NewUserRepository(s.T())
}

func makeUsers(count int) []*entities.User {
	users := make([]*entities.User, count)
	for i := range users {
		name, _ := vos.NewUserName("User")
		email, _ := vos.NewEmail("user@example.com")
		u, _ := entities.NewUser(name, email)
		users[i] = u
	}
	return users
}

func (s *ListUsersUseCaseSuite) TestExecute() {
	scenarios := []struct {
		name   string
		input  ListUsersInput
		setup  func()
		expect func(output *ListUsersOutput, err error)
	}{
		{
			name:  "deve retornar lista sem cursor",
			input: ListUsersInput{Limit: 20, Cursor: ""},
			setup: func() {
				users := makeUsers(3)
				s.userRepository.EXPECT().
					FindAll(s.ctx, 20, "").
					Return(users, nil, nil).
					Once()
			},
			expect: func(output *ListUsersOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Users, 3)
				s.Nil(output.NextCursor)
			},
		},
		{
			name:  "deve retornar lista com próximo cursor",
			input: ListUsersInput{Limit: 2, Cursor: ""},
			setup: func() {
				users := makeUsers(2)
				cursor := "next-cursor-value"
				s.userRepository.EXPECT().
					FindAll(s.ctx, 2, "").
					Return(users, &cursor, nil).
					Once()
			},
			expect: func(output *ListUsersOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Users, 2)
				s.NotNil(output.NextCursor)
			},
		},
		{
			name:  "deve retornar lista vazia",
			input: ListUsersInput{Limit: 20, Cursor: ""},
			setup: func() {
				s.userRepository.EXPECT().
					FindAll(s.ctx, 20, "").
					Return([]*entities.User{}, nil, nil).
					Once()
			},
			expect: func(output *ListUsersOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Empty(output.Users)
				s.Nil(output.NextCursor)
			},
		},
		{
			name:  "deve retornar erro ao falhar no repositório",
			input: ListUsersInput{Limit: 20, Cursor: ""},
			setup: func() {
				s.userRepository.EXPECT().
					FindAll(s.ctx, 20, "").
					Return(nil, nil, errors.New("db error")).
					Once()
			},
			expect: func(output *ListUsersOutput, err error) {
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
			uc := NewListUsersUseCase(obs, fm, s.userRepository)
			output, err := uc.Execute(s.ctx, scenario.input)

			scenario.expect(output, err)
		})
	}
}
