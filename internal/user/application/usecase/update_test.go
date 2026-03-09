package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	devkitVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	userdomain "github.com/jailtonjunior94/financial/internal/user/domain"
	repositoryMock "github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateUserUseCaseSuite struct {
	suite.Suite
	ctx            context.Context
	userRepository *repositoryMock.UserRepository
	hash           encrypt.HashAdapter
}

func TestUpdateUserUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateUserUseCaseSuite))
}

func (s *UpdateUserUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepository = repositoryMock.NewUserRepository(s.T())
	s.hash = encrypt.NewHashAdapter()
}

func strPtr(s string) *string { return &s }

type errHashAdapter struct{}

func (e *errHashAdapter) GenerateHash(_ string) (string, error) {
	return "", errors.New("bcrypt error")
}

func (e *errHashAdapter) CheckHash(_, _ string) bool { return false }

func (s *UpdateUserUseCaseSuite) TestExecute() {
	type args struct {
		id           string
		input        *dtos.UpdateUserInput
		overrideHash encrypt.HashAdapter
	}

	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(output interface{}, err error)
	}{
		{
			name: "deve atualizar nome com sucesso",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Name: strPtr("New Name")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					Update(s.ctx, mock.AnythingOfType("*entities.User")).
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.NoError(err)
				s.NotNil(output)
			},
		},
		{
			name: "deve atualizar senha com sucesso",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Password: strPtr("newSecurePassword123")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					Update(s.ctx, mock.AnythingOfType("*entities.User")).
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.NoError(err)
				s.NotNil(output)
			},
		},
		{
			name: "deve atualizar email para o mesmo email do proprio usuario com sucesso",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Email: strPtr("john@example.com")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				id, _ := devkitVos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
				user.SetID(id)
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					FindByEmail(s.ctx, "john@example.com").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					Update(s.ctx, mock.AnythingOfType("*entities.User")).
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.NoError(err)
				s.NotNil(output)
			},
		},
		{
			name: "deve retornar ErrEmailAlreadyExists ao atualizar email duplicado",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Email: strPtr("existing@example.com")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				otherUser := newTestUserEntity(s.T())

				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					FindByEmail(s.ctx, "existing@example.com").
					Return(otherUser, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, userdomain.ErrEmailAlreadyExists)
			},
		},
		{
			name: "deve retornar ErrUserNotFound quando usuário não existe",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Name: strPtr("New Name")},
			},
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
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Name: strPtr("New Name")},
			},
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
		{
			name: "deve retornar erro ao atualizar nome invalido (vazio)",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Name: strPtr("")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "deve retornar erro ao atualizar email invalido",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Email: strPtr("not-a-valid-email")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "deve retornar erro ao falhar no hash de senha",
			args: args{
				id:           "550e8400-e29b-41d4-a716-446655440000",
				input:        &dtos.UpdateUserInput{Password: strPtr("newSecurePassword123")},
				overrideHash: &errHashAdapter{},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "bcrypt error")
			},
		},
		{
			name: "partial update (só nome) — email e password permanecem inalterados",
			args: args{
				id:    "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.UpdateUserInput{Name: strPtr("Updated Name")},
			},
			setup: func() {
				user := newTestUserEntity(s.T())
				s.userRepository.EXPECT().
					FindByID(s.ctx, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil).
					Once()
				s.userRepository.EXPECT().
					Update(s.ctx, mock.AnythingOfType("*entities.User")).
					Return(user, nil).
					Once()
			},
			expect: func(output interface{}, err error) {
				s.NoError(err)
				s.NotNil(output)
				typed, ok := output.(*dtos.UserOutput)
				s.True(ok)
				s.Equal("john@example.com", typed.Email)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.setup()

			obs := fake.NewProvider()
			fm := metrics.NewTestFinancialMetrics()
			hash := encrypt.HashAdapter(s.hash)
			if scenario.args.overrideHash != nil {
				hash = scenario.args.overrideHash
			}
			uc := NewUpdateUserUseCase(obs, fm, hash, s.userRepository)
			output, err := uc.Execute(s.ctx, scenario.args.id, scenario.args.input)

			scenario.expect(output, err)
		})
	}
}
