package user

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type UserModule struct {
	UserRouter *http.UserRouter
}

func NewUserModule(db *sql.DB, cfg *configs.Config, o11y observability.Observability) UserModule {
	hash := encrypt.NewHashAdapter()
	jwt := auth.NewJwtAdapter(cfg, o11y)
	errorHandler := httperrors.NewErrorHandler(o11y)

	financialMetrics := metrics.NewFinancialMetrics(o11y)
	userRepository := repositories.NewUserRepository(db, o11y, financialMetrics)

	authUseCase := usecase.NewTokenUseCase(cfg, o11y, hash, jwt, userRepository)
	createUserUseCase := usecase.NewCreateUserUseCase(o11y, hash, userRepository)

	authHandler := http.NewAuthHandler(o11y, errorHandler, authUseCase)
	userHandler := http.NewUserHandler(o11y, errorHandler, createUserUseCase)

	userRouter := http.NewUserRouter(authHandler, userHandler)

	return UserModule{UserRouter: userRouter}
}
