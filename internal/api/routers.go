package api

import (
	"net/http"

	"go-form-hub/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Route struct {
	Name         string
	Method       string
	Path         string
	Handler      http.HandlerFunc
	AuthRequired bool
}

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
func NewRouter(cfg *config.Config, authMiddleware, currentUserMiddleware, csrfMiddleware func(http.HandlerFunc) http.HandlerFunc, routers ...Router) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  AllowOriginFunc,
		AllowedOrigins:   []string{cfg.AllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Csrf-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	for _, api := range routers {
		for _, route := range api.Routes() {
			handler := route.Handler
			if route.AuthRequired {
				handler = csrfMiddleware(handler)
				handler = authMiddleware(handler)
			} else {
				handler = currentUserMiddleware(handler)
			}

			apiPath := "/api/v1" + route.Path
			router.Method(route.Method, apiPath, handler)
		}
	}

	return router
}

func AllowOriginFunc(_ *http.Request, _ string) bool {
	return true
}
