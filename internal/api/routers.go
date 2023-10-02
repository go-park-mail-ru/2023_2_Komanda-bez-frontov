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

// NewRouter creates a new chi router and adds the provided routers to it.
//
// routers: A variadic parameter of type `Router` representing the routers to be added to the chi router.
// Returns: A chi router with the added routers.
func NewRouter(routers ...Router) chi.Router {
	router := chi.NewRouter()
	for _, api := range routers {
		for _, route := range api.Routes() {
			// TODO: add auth & recovery middleware
			router.Method(route.Method, route.Path, route.Handler)
		}
	}

	return router
}
