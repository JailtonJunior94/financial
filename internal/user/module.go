package user

import (
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repository"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/web"
	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
)

func RegisterAuthModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repository.NewUserRepository(ioc.DB, ioc.Observability)
	authUseCase := usecase.NewTokenUseCase(ioc.Config, ioc.Logger, ioc.Hash, ioc.Jwt, userRepository, ioc.Observability)
	authHandler := web.NewAuthHandler(ioc.Observability, authUseCase)
	web.NewAuthRoute(router, web.WithTokenHandler(authHandler.Token))
}

func RegisterUserModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repository.NewUserRepository(ioc.DB, ioc.Observability)
	createUserUseCase := usecase.NewCreateUserUseCase(ioc.Logger, ioc.Hash, userRepository)
	userHandler := web.NewUserHandler(createUserUseCase)
	web.NewUserRoute(router, web.WithCreateUserHandler(userHandler.Create))
}
