package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Route struct defines parameters of API endpoint
type Route struct {
	Name         string
	Method       string
	Path         string
	Handler      http.HandlerFunc
	AuthRequired bool
}

// Router defines a list of routes of API
type Router interface {
	Routes() []Route
}

// NewRouter creates a new router with the given authentication middleware and routers.
//
// The `authMiddleware` parameter is a function that takes a `http.HandlerFunc` and returns
// a new `http.HandlerFunc`. It is used to authenticate requests before they are handled
// by the router.
//
// The `routers` parameter is a variadic parameter that allows you to pass one or more
// `Router` objects. These routers contain the routes that will be added to the main router.
//
// The function returns a new `chi.Router` that has all the routes from the provided routers
// added to it.
func NewRouter(authMiddleware func(http.HandlerFunc) http.HandlerFunc, routers ...Router) chi.Router {
	router := chi.NewRouter()
	for _, api := range routers {
		for _, route := range api.Routes() {
			handler := route.Handler
			if route.AuthRequired {
				handler = authMiddleware(handler)
			}

			apiPath := "/api/v1" + route.Path
			router.Method(route.Method, apiPath, handler)
		}
	}

	return router
}
