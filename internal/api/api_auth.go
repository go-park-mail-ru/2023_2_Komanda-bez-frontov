package api

import (
	"encoding/json"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/services/auth"
	resp "go-form-hub/internal/services/service_response"
	"io"
	"net/http"
	"strconv"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type AuthAPIController struct {
	authService      auth.Service
	errorHandler     ErrorHandler
	validator        *validator.Validate
	cookieExpiration time.Duration
}

func NewAuthAPIController(authService auth.Service, v *validator.Validate, cookieExpiration time.Duration) Router {
	return &AuthAPIController{
		authService:      authService,
		errorHandler:     HandleError,
		validator:        v,
		cookieExpiration: cookieExpiration,
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
			Method:       http.MethodDelete,
			Path:         "/logout",
			Handler:      c.Logout,
			AuthRequired: true,
		},
	}
}

// nolint:dupl
func (c *AuthAPIController) Login(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == nil {

		sessionID, err := strconv.ParseInt(session.Value, 10, 64)
		if err != nil {
			c.errorHandler(w, err, &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}

		isValid, err := c.authService.IsSessionValid(r.Context(), sessionID)
		if err != nil {
			c.errorHandler(w, err, &resp.Response{Body: nil, StatusCode: http.StatusInternalServerError})
			return
		}
		if isValid {
			c.errorHandler(w, fmt.Errorf("already logged in"), &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}
	}

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
		Value:   fmt.Sprintf("%d", sessionID),
		Expires: time.Now().Add(c.cookieExpiration),
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *AuthAPIController) Signup(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == nil {
		sessionID, err := strconv.ParseInt(session.Value, 10, 64)
		if err != nil {
			c.errorHandler(w, err, &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}

		isValid, err := c.authService.IsSessionValid(r.Context(), sessionID)
		if err != nil {
			c.errorHandler(w, err, &resp.Response{Body: nil, StatusCode: http.StatusInternalServerError})
			return
		}
		if isValid {
			c.errorHandler(w, fmt.Errorf("already logged in"), &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("api_auth read_body err: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	var user model.UserSignUp
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		log.Error().Msgf("api_auth unmarshal err: %e", err)
		c.errorHandler(w, err, nil)
		return
	}

	result, sessionID, err := c.authService.AuthSignUp(r.Context(), &user)
	if err != nil {
		log.Error().Msgf("api_auth sugnip err: %e", err)
		c.errorHandler(w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   fmt.Sprintf("%d", sessionID),
		Expires: time.Now().Add(c.cookieExpiration),
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *AuthAPIController) Logout(w http.ResponseWriter, r *http.Request) {
	result, _, err := c.authService.AuthLogout(r.Context())
	if err != nil {
		c.errorHandler(w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}
	http.SetCookie(w, cookie)
	EncodeJSONResponse(result.Body, result.StatusCode, w)
}
