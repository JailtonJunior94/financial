package auth

import (
	"testing"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/domain/user/entity"
	repositoryMock "github.com/jailtonjunior94/financial/internal/infrastructure/user/repository/mock"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TokenSuite struct {
	suite.Suite

	logger logger.Logger
	config *configs.Config
	hash   encrypt.HashAdapter
	jwt    authentication.JwtAdapter
}

func TestTokenSuite(t *testing.T) {
	suite.Run(t, new(TokenSuite))
}

func (s *TokenSuite) SetupTest() {
	s.config = &configs.Config{
		AuthExpirationAt: 8,
		AuthSecretKey:    "my_secret_key",
	}
	s.logger = logger.NewLogger()
	s.jwt = authentication.NewJwtAdapter(s.logger, s.config)
	s.hash = encrypt.NewHashAdapter()
}

func (s *TokenSuite) TestToken() {
	type (
		args struct {
			input *AuthInput
		}
		fields struct {
			userRepository *repositoryMock.UserRepository
		}
	)

	passwordHash, _ := s.hash.GenerateHash("my_password@2024")
	user, _ := entity.NewUser("John Mckinley", "john.mckinley@examplepetstore.com")
	user.SetPassword(passwordHash)

	scenarios := []struct {
		name     string
		args     args
		fields   fields
		expected func(res *AuthOutput, err error)
	}{
		{
			name: "must return a token when username and password are valid",
			args: args{input: &AuthInput{Email: "john.mckinley@examplepetstore.com", Password: "my_password@2024"}},
			fields: fields{
				userRepository: func() *repositoryMock.UserRepository {
					userRepository := &repositoryMock.UserRepository{}
					userRepository.
						On("FindByEmail", mock.Anything).
						Return(user, nil)
					return userRepository
				}(),
			},
			expected: func(res *AuthOutput, err error) {
				s.NoError(err)
				s.NotNil(res)
			},
		},
	}

	for _, scenario := range scenarios {
		s.T().Run(scenario.name, func(t *testing.T) {
			tokenUseCase := NewTokenUseCase(s.config, s.logger, s.hash, s.jwt, scenario.fields.userRepository)
			token, err := tokenUseCase.Execute(scenario.args.input)
			scenario.expected(token, err)
		})
	}
}
