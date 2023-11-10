package usecase

import (
	"time"

	"github.com/jailtonjunior94/financial/internal/domain/user/entity"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
)

type CreateUserUseCase interface {
	Execute(input *CreateUserInput) (*CreateUserOutput, error)
}

type createUserUseCase struct {
	Repository interfaces.UserRepository
}

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserOutput struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewCreateUserUseCase(repository interfaces.UserRepository) CreateUserUseCase {
	return &createUserUseCase{Repository: repository}
}

func (u *createUserUseCase) Execute(input *CreateUserInput) (*CreateUserOutput, error) {
	newUser, err := entity.NewUser(input.Name, input.Email)
	if err != nil {
		return nil, err
	}
	newUser.SetPassword(input.Password)

	user, err := u.Repository.Create(newUser)
	if err != nil {
		return nil, err
	}

	return &CreateUserOutput{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
