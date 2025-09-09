package usecase

import (
	"context"

	"testing"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/auth"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TokenSuite struct {
	suite.Suite

	ctx        context.Context
	config     *configs.Config
	jwt        auth.JwtAdapter
	o11y       o11y.Observability
	hash       encrypt.HashAdapter
	repository *repositoryMock.UserRepository
}

func TestTokenSuite(t *testing.T) {
	suite.Run(t, new(TokenSuite))
}

func (s *TokenSuite) SetupTest() {
	s.ctx = context.Background()
	s.config = &configs.Config{
		AuthConfig: configs.AuthConfig{
			AuthTokenDuration: 60,
			AuthSecretKey:     "your_secret_key",
		},
	}

	s.o11y = o11y.NewDevelopmentObservability("test", "1.0.0")
	s.hash = encrypt.NewHashAdapter()
	s.jwt = auth.NewJwtAdapter(s.config, s.o11y)
	s.repository = repositoryMock.NewUserRepository(s.T())
}

func (s *TokenSuite) TearDownTest() {
	s.repository.AssertExpectations(s.T())
}

func (s *TokenSuite) TestToken() {
	type (
		args struct {
			input *dtos.AuthInput
		}
		fields struct {
			repository *repositoryMock.UserRepository
		}
	)

	passwordHash, _ := s.hash.GenerateHash("my_password@2024")
	user, err := factories.CreateUser("John Mckinley", "john.mckinley@examplepetstore.com")
	s.Require().NoError(err)
	err = user.SetPassword(passwordHash)
	s.Require().NoError(err)

	scenarios := []struct {
		name     string
		args     args
		fields   fields
		expected func(res *dtos.AuthOutput, err error)
	}{
		{
			name: "must return a token when username and password are valid",
			args: args{input: &dtos.AuthInput{Email: "john.mckinley@examplepetstore.com", Password: "my_password@2024"}},
			fields: fields{
				repository: func() *repositoryMock.UserRepository {
					s.repository.
						EXPECT().
						FindByEmail(mock.Anything, mock.Anything).
						Return(user, nil).
						Once()
					return s.repository
				}(),
			},
			expected: func(res *dtos.AuthOutput, err error) {
				s.NoError(err)
				s.NotNil(res)
			},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			tokenUseCase := NewTokenUseCase(s.config, s.o11y, s.hash, s.jwt, scenario.fields.repository)
			token, err := tokenUseCase.Execute(s.ctx, scenario.args.input)
			scenario.expected(token, err)
		})
	}
}
