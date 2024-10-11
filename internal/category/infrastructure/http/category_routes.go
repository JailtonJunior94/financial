package http

import (
	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

type categoryRoutes struct {
	routes []httpserver.Route
}

func NewCategoryRoutes() *categoryRoutes {
	return &categoryRoutes{}
}

func (u *categoryRoutes) Register(route httpserver.Route) {
	u.routes = append(u.routes, route)
}

func (u *categoryRoutes) Routes() []httpserver.Route {
	return u.routes
}
