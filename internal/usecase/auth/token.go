package auth

import (
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
)

type TokenUseCase interface {
	Execute(input *AuthInput) (*AuthOutput, error)
}

type tokenUseCase struct {
	Hash       encrypt.HashAdapter
	Jwt        authentication.JwtAdapter
	Repository interfaces.UserRepository
}

func NewTokenUseCase(hash encrypt.HashAdapter, jwt authentication.JwtAdapter, repository interfaces.UserRepository) TokenUseCase {
	return &tokenUseCase{Hash: hash, Jwt: jwt, Repository: repository}
}

func (u *tokenUseCase) Execute(input *AuthInput) (*AuthOutput, error) {
	user, err := u.Repository.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	if !u.Hash.CheckHash(user.Password, input.Password) {
		return nil, nil
	}

	token, err := u.Jwt.GenerateTokenJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &AuthOutput{Token: token}, nil
}
