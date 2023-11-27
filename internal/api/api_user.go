package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"go-form-hub/internal/model"
	"go-form-hub/microservices/user/profile"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type UserAPIController struct {
	service         profile.ProfileClient
	validator       *validator.Validate
	responseEncoder ResponseEncoder
}

func NewUserAPIController(service profile.ProfileClient, v *validator.Validate, responseEncoder ResponseEncoder) Router {
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
	userID := &profile.CurrentUserID{
		Id: currentUser.ID,
	}

	result, err := c.service.UserGet(ctx, userID)
	if err != nil {
		log.Error().Msgf("api_user profile_get err: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
	}

	var user *profile.User
	err = result.Body.UnmarshalTo(user)
	if err != nil {
		log.Error().Msgf("couldnt parse response: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	modelUser := &model.UserGet{
		ID:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Avatar:    &user.Avatar,
		Username:  user.Username,
	}
	c.responseEncoder.EncodeJSONResponse(ctx, modelUser, int(result.Code), w)
}

func (c *UserAPIController) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		log.Error().Msgf("user_api user_update read_body error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var updatedUser model.UserUpdate
	if err = json.Unmarshal(requestJSON, &updatedUser); err != nil {
		log.Error().Msgf("user_api user_update unmarshal error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	userUpdate := profile.UserUpdate{
		Avatar:      *updatedUser.Avatar,
		Username:    updatedUser.Username,
		FirstName:   updatedUser.FirstName,
		LastName:    updatedUser.LastName,
		Password:    updatedUser.Password,
		NewPassword: updatedUser.NewPassword,
		Email:       updatedUser.Email,
	}

	result, err := c.service.Update(ctx, &userUpdate)
	if err != nil {
		log.Error().Msgf("user_api user_update error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var user *profile.User
	err = result.Body.UnmarshalTo(user)
	if err != nil {
		log.Error().Msgf("couldnt parse response: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	modelUser := &model.UserGet{
		ID:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Avatar:    &user.Avatar,
		Username:  user.Username,
	}

	c.responseEncoder.EncodeJSONResponse(ctx, modelUser, int(result.Code), w)
}

func (c *UserAPIController) UserAvatarGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username, err := url.PathUnescape(chi.URLParam(r, "username"))
	if err != nil {
		log.Error().Msgf("user_api user_avatar_get unescape error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}
	usernameMsg := &profile.CurrentUserUsername{
		Username: username,
	}

	result, err := c.service.AvatarGet(ctx, usernameMsg)
	if err != nil {
		log.Error().Msgf("user_api user_avatar_get error: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	var avatar *profile.UserAvatar
	err = result.Body.UnmarshalTo(avatar)
	if err != nil {
		log.Error().Msgf("couldnt parse response: %v", err)
		c.responseEncoder.HandleError(ctx, w, err, nil)
		return
	}

	avatarGet := &model.UserAvatarGet{
		Username: avatar.Username,
		Avatar:   &avatar.Avatar,
	}

	c.responseEncoder.EncodeJSONResponse(ctx, avatarGet, int(result.Code), w)
}

//TODO: в Handle error завернуть response.code
