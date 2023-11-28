package controller

import (
	"context"

	"go-form-hub/internal/model"
	"go-form-hub/microservices/user/profile"
	"go-form-hub/microservices/user/usecase"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/anypb"
)

const defaultAvatar = ""

type Controller struct {
	profile.UnimplementedProfileServer

	usecase   usecase.UserUseCase
	validator *validator.Validate
}

func NewController(userService usecase.UserUseCase, v *validator.Validate) *Controller {
	return &Controller{
		usecase:   userService,
		validator: v,
	}
}

func (pm *Controller) UserGet(ctx context.Context, userID *profile.CurrentUserID) (*profile.Response, error) {
	result, err := pm.usecase.UserGet(ctx, userID.Id)
	if err != nil {
		return nil, err
	}

	userGet := result.Body.(*model.UserGet)
	if userGet.Avatar == nil {
		empty := defaultAvatar
		userGet.Avatar = &empty
	}
	userMsg := &profile.User{
		Email:     userGet.Email,
		FirstName: userGet.FirstName,
		LastName:  userGet.LastName,
		Username:  userGet.Username,
		Id:        userGet.ID,
		Avatar:    *userGet.Avatar,
	}
	body, err := anypb.New(userMsg)
	if err != nil {
		return nil, err
	}

	response := &profile.Response{
		Code: int64(result.StatusCode),
		Body: body,
	}

	return response, nil
}
func (pm *Controller) AvatarGet(ctx context.Context, userID *profile.CurrentUserUsername) (*profile.Response, error) {
	result, err := pm.usecase.UserGetAvatar(ctx, userID.Username)
	if err != nil {
		return nil, err
	}

	userGet := result.Body.(*model.UserAvatarGet)
	if userGet.Avatar == nil {
		empty := defaultAvatar
		userGet.Avatar = &empty
	}
	userMsg := &profile.UserAvatar{
		Username: userGet.Username,
		Avatar:   *userGet.Avatar,
	}
	body, err := anypb.New(userMsg)
	if err != nil {
		return nil, err
	}

	response := &profile.Response{
		Code: int64(result.StatusCode),
		Body: body,
	}

	return response, nil
}
func (pm *Controller) Update(ctx context.Context, userUpdate *profile.UserUpdateReq) (*profile.Response, error) {
	userModelUpdate := &model.UserUpdate{
		Username:    userUpdate.Update.Username,
		FirstName:   userUpdate.Update.FirstName,
		LastName:    userUpdate.Update.LastName,
		Password:    userUpdate.Update.Password,
		NewPassword: userUpdate.Update.NewPassword,
		Email:       userUpdate.Update.Email,
		Avatar:      &userUpdate.Update.Avatar,
	}

	ctx = context.WithValue(ctx, model.ContextCurrentUser, &model.UserGet{
		ID:        userUpdate.CurrentUser.Id,
		Username:  userUpdate.CurrentUser.Username,
		FirstName: userUpdate.CurrentUser.FirstName,
		LastName:  userUpdate.CurrentUser.LastName,
		Email:     userUpdate.CurrentUser.Email,
	})

	result, err := pm.usecase.UserUpdate(ctx, userModelUpdate)
	if err != nil {
		return nil, err
	}

	userGet := result.Body.(*model.UserGet)
	if userGet.Avatar == nil {
		empty := defaultAvatar
		userGet.Avatar = &empty
	}
	userMsg := &profile.User{
		Email:     userGet.Email,
		FirstName: userGet.FirstName,
		LastName:  userGet.LastName,
		Username:  userGet.Username,
		Id:        userGet.ID,
		Avatar:    *userGet.Avatar,
	}
	body, err := anypb.New(userMsg)
	if err != nil {
		return nil, err
	}

	response := &profile.Response{
		Code: int64(result.StatusCode),
		Body: body,
	}

	return response, nil
}
