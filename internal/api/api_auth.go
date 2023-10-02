package api

import (
	"github.com/go-playground/validator/v10"
)

type UserAPIController struct {
	//service   user.Service
	validator *validator.Validate
}

// func NewUserAPIController(service user.Service, v *validator.Validate) Router {
// 	return &FormAPIController{
// 		service:   service,
// 		validator: v,
// 	}
// }
