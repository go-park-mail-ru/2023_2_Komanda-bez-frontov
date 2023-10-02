package api

import (
	// "encoding/json"
	// "go-form-hub/internal/model"
	// "go-form-hub/internal/services/form"
	// "io"
	"net/http"
	// "net/url"

	// "github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type UserAPIController struct {
	//service   user.Service
	validator *validator.Validate
}

// func NewUserAPIControllr(service user.Service, v *validator.Validate) Router {
// 	return &FormAPIController{
// 		service:   service,
// 		validator: v,
// 	}
// }

func (c *UserAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "Login",
			Method:       http.MethodPost,
			Path:         "/login",
			//Handler:      c.FormSave,
			AuthRequired: false,
		},
		{
			Name:         "Signup",
			Method:       http.MethodPost,
			Path:         "/signup",
			//Handler:      c.FormList,
			AuthRequired: false,
		},
	}
}