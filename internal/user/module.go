package user

import (
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repository"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/rest"
	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
)

func RegisterAuthModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repository.NewUserRepository(ioc.DB, ioc.Observability)
	authUseCase := usecase.NewTokenUseCase(ioc.Config, ioc.Observability, ioc.Hash, ioc.Jwt, userRepository)
	authHandler := rest.NewAuthHandler(ioc.Observability, authUseCase)
	rest.NewAuthRoute(router, rest.WithTokenHandler(authHandler.Token))
}

func RegisterUserModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repository.NewUserRepository(ioc.DB, ioc.Observability)
	createUserUseCase := usecase.NewCreateUserUseCase(ioc.Observability, ioc.Hash, userRepository)
	userHandler := rest.NewUserHandler(ioc.Observability, createUserUseCase)
	rest.NewUserRoute(router, rest.WithCreateUserHandler(userHandler.Create))
}
