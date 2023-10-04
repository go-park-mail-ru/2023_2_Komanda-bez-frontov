package api

import (
	"encoding/json"
	"go-form-hub/internal/model"
	"go-form-hub/internal/services/auth"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type AuthAPIController struct {
	authService  auth.Service
	errorHandler ErrorHandler
	validator    *validator.Validate
}

func NewAuthAPIController(authService auth.Service, v *validator.Validate) Router {
	return &AuthAPIController{
		authService:  authService,
		errorHandler: HandleError,
		validator:    v,
	}
}

func (c *AuthAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "Login",
			Method:       http.MethodPost,
			Path:         "/login",
			Handler:      c.Login,
			AuthRequired: false,
		},
		{
			Name:         "Signup",
			Method:       http.MethodPost,
			Path:         "/signup",
			Handler:      c.Signup,
			AuthRequired: false,
		},
		{
			Name:         "Logout",
			Method:       http.MethodPost,
			Path:         "/logout",
			Handler:      c.Logout,
			AuthRequired: true,
		},
	}
}

// nolint:dupl
func (c *AuthAPIController) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: redirect authenticated user to home

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		c.errorHandler(w, err, nil)
		return
	}

	var user model.UserLogin
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		c.errorHandler(w, err, nil)
		return
	}

	result, sessionID, err := c.authService.AuthLogin(r.Context(), &user)
	if err != nil {
		c.errorHandler(w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *AuthAPIController) Signup(w http.ResponseWriter, r *http.Request) {
	// TODO: redirect authenticated user to home

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		c.errorHandler(w, err, nil)
		return
	}

	var user model.UserSignUp
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		c.errorHandler(w, err, nil)
		return
	}

	result, sessionID, err := c.authService.AuthSignUp(r.Context(), &user)
	if err != nil {
		c.errorHandler(w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *AuthAPIController) Logout(w http.ResponseWriter, r *http.Request) {
	result, sessionID, err := c.authService.AuthLogout(r.Context())
	if err != nil {
		c.errorHandler(w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}
