package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/auth"
	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type AuthAPIController struct {
	authService      auth.Service
	validator        *validator.Validate
	cookieExpiration time.Duration
	responseEncoder  ResponseEncoder
}

func NewAuthAPIController(authService auth.Service, v *validator.Validate, cookieExpiration time.Duration, responseEncoder ResponseEncoder) Router {
	return &AuthAPIController{
		authService:      authService,
		validator:        v,
		cookieExpiration: cookieExpiration,
		responseEncoder:  responseEncoder,
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
		{
			Name:         "IsAuthorized",
			Method:       http.MethodGet,
			Path:         "/is_authorized",
			Handler:      c.IsAuthorized,
			AuthRequired: true,
		},
	}
}

// nolint:dupl
func (c *AuthAPIController) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, err := r.Cookie("session_id")
	if err == nil {
		isValid, err := c.authService.IsSessionValid(ctx, session.Value)
		if err != nil {
			c.responseEncoder.HandleError(ctx, w, err, &resp.Response{Body: nil, StatusCode: http.StatusInternalServerError})
			return
		}
		if isValid {
			c.responseEncoder.HandleError(ctx, w, fmt.Errorf("already logged in"), &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var user model.UserLogin
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, sessionID, err := c.authService.AuthLogin(ctx, &user)
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(c.cookieExpiration),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

// nolint:dupl
func (c *AuthAPIController) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, err := r.Cookie("session_id")
	if err == nil {
		isValid, err := c.authService.IsSessionValid(ctx, session.Value)
		if err != nil {
			c.responseEncoder.HandleError(ctx, w, err, &resp.Response{Body: nil, StatusCode: http.StatusInternalServerError})
			return
		}
		if isValid {
			c.responseEncoder.HandleError(ctx, w, fmt.Errorf("already logged in"), &resp.Response{Body: nil, StatusCode: http.StatusBadRequest})
			return
		}
	}

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("api_auth read_body err: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var user model.UserSignUp
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		log.Error().Msgf("api_auth unmarshal err: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, sessionID, err := c.authService.AuthSignUp(ctx, &user)
	if err != nil {
		log.Error().Msgf("api_auth sugnip err: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(c.cookieExpiration),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *AuthAPIController) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, _, err := c.authService.AuthLogout(ctx)
	if err != nil {
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}
	http.SetCookie(w, cookie)
	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *AuthAPIController) IsAuthorized(w http.ResponseWriter, r *http.Request) {
	c.responseEncoder.EncodeJSONResponse(r.Context(), nil, http.StatusOK, w)
}
