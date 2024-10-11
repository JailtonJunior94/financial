package http

import (
	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

type userRoutes struct {
	routes []httpserver.Route
}

func NewUserRoutes() *userRoutes {
	return &userRoutes{}
}

func (u *userRoutes) Register(route httpserver.Route) {
	u.routes = append(u.routes, route)
}

func (u *userRoutes) Routes() []httpserver.Route {
	return u.routes
}
