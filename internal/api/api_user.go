package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/user"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type UserAPIController struct {
	service         user.Service
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewUserAPIController(service user.Service, v *validator.Validate, responseEncoder ResponseEncoder) Router {
	return &UserAPIController{
		service:         service,
		validator:       v,
		responseEncoder: responseEncoder,
	}
}

func (c *UserAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "Profile",
			Method:       http.MethodGet,
			Path:         "/profile",
			Handler:      c.ProfileGet,
			AuthRequired: true,
		},
		{
			Name:         "ProfileUpdate",
			Method:       http.MethodPut,
			Path:         "/profile/update",
			Handler:      c.ProfileUpdate,
			AuthRequired: true,
		},
		{
			Name:         "UserAvatarGet",
			Method:       http.MethodGet,
			Path:         "/user/{username}/avatar",
			Handler:      c.UserAvatarGet,
			AuthRequired: false,
		},
	}
}

func (c *UserAPIController) ProfileGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	result, err := c.service.UserGet(ctx, currentUser.ID)
	if err != nil {
		log.Error().Msgf("api_user profile_get err: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *UserAPIController) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("user_api user_update read_body error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var updatedUser model.UserUpdate
	if err = json.Unmarshal(requestJSON, &updatedUser); err != nil {
		log.Error().Msgf("user_api user_update unmarshal error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.UserUpdate(ctx, &updatedUser)
	if err != nil {
		log.Error().Msgf("user_api user_update error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}

func (c *UserAPIController) UserAvatarGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username, err := url.PathUnescape(chi.URLParam(r, "username"))
	if err != nil {
		log.Error().Msgf("user_api user_avatar_get unescape error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	result, err := c.service.UserGetAvatar(ctx, username)
	if err != nil {
		log.Error().Msgf("user_api user_avatar_get error: %e", err)
		c.responseEncoder.HandleError(ctx, w, err, result)
		return
	}

	c.responseEncoder.EncodeJSONResponse(ctx, result.Body, result.StatusCode, w)
}
