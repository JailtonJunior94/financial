package user

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
)

type UserModule struct {
	UserRouter *http.UserRouter
}

func NewUserModule(db *sql.DB, cfg *configs.Config, o11y observability.Observability) UserModule {
	userRepository := repositories.NewUserRepository(db, o11y)

	jwt := auth.NewJwtAdapter(cfg, o11y)
	hash := encrypt.NewHashAdapter()

	authUseCase := usecase.NewTokenUseCase(cfg, o11y, hash, jwt, userRepository)
	createUserUseCase := usecase.NewCreateUserUseCase(o11y, hash, userRepository)

	authHandler := http.NewAuthHandler(o11y, authUseCase)
	userHandler := http.NewUserHandler(o11y, createUserUseCase)

	userRouter := http.NewUserRouter(authHandler, userHandler)

	return UserModule{
		UserRouter: userRouter,
	}
}
