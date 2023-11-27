package profile

import (
	"context"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/user"
	"go-form-hub/microservices/user/profile"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/anypb"
)

const defaultAvatar = ""

type ProfileController struct {
	profile.UnimplementedProfileServer

	service   user.Service
	validator *validator.Validate
}

func NewProfileController(userService user.Service, v *validator.Validate) *ProfileController {
	return &ProfileController{
		service:   userService,
		validator: v,
	}
}

func (pm *ProfileController) UserGet(ctx context.Context, userID *profile.CurrentUserID) (*profile.Response, error) {
	result, err := pm.service.UserGet(ctx, userID.Id)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserGet)
	if user.Avatar == nil {
		empty := defaultAvatar
		user.Avatar = &empty
	}
	userMsg := &profile.User{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Id:        user.ID,
		Avatar:    *user.Avatar,
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
func (pm *ProfileController) AvatarGet(ctx context.Context, userID *profile.CurrentUserUsername) (*profile.Response, error) {
	result, err := pm.service.UserGetAvatar(ctx, userID.Username)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserAvatarGet)
	if user.Avatar == nil {
		empty := defaultAvatar
		user.Avatar = &empty
	}
	userMsg := &profile.UserAvatar{
		Username: user.Username,
		Avatar:   *user.Avatar,
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
func (pm *ProfileController) Update(ctx context.Context, userUpdate *profile.UserUpdateReq) (*profile.Response, error) {
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

	result, err := pm.service.UserUpdate(ctx, userModelUpdate)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserGet)
	if user.Avatar == nil {
		empty := defaultAvatar
		user.Avatar = &empty
	}
	userMsg := &profile.User{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Id:        user.ID,
		Avatar:    *user.Avatar,
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
