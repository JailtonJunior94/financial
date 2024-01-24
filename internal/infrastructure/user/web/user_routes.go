package web

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/infrastructure/web/middlewares"

	"github.com/go-chi/chi/v5"
)

type Routes func(userRoute *userRoute)
type userRoute struct {
	Authorization     middlewares.Authorization
	CreateUserHandler func(w http.ResponseWriter, r *http.Request)
}

func NewUserRoutes(router *chi.Mux, middleware middlewares.Authorization, userRoutes ...Routes) *userRoute {
	route := &userRoute{}
	for _, userRoute := range userRoutes {
		userRoute(route)
	}
	route.Register(middleware, router)
	return route
}

func (u *userRoute) Register(middleware middlewares.Authorization, router *chi.Mux) {
	router.Route("/api/v1/users", func(r chi.Router) {
		r.Post("/", u.CreateUserHandler)
	})
}

func WithCreateUserHandler(handler func(w http.ResponseWriter, r *http.Request)) Routes {
	return func(userRouter *userRoute) {
		userRouter.CreateUserHandler = handler
	}
}
