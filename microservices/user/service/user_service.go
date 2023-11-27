package profile

import (
	"context"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/user"
	"go-form-hub/microservices/user/profile"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/anypb"
)

type ProfileManager struct {
	service   user.Service
	validator *validator.Validate
}

func NewProfileManager(userService user.Service, v *validator.Validate) *ProfileManager {
	return &ProfileManager{
		service:   userService,
		validator: v,
	}
}

func (pm *ProfileManager) UserGet(ctx context.Context, userID *profile.CurrentUserID) (*profile.Response, error) {
	result, err := pm.service.UserGet(ctx, userID.Id)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserGet)
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
func (pm *ProfileManager) AvatarGet(ctx context.Context, userID *profile.CurrentUserUsername) (*profile.Response, error) {
	result, err := pm.service.UserGetAvatar(ctx, userID.Username)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserAvatarGet)
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
func (pm *ProfileManager) Update(ctx context.Context, userUpdate *profile.UserUpdate) (*profile.Response, error) {
	userModelUpdate := &model.UserUpdate{
		Username:    userUpdate.Username,
		FirstName:   userUpdate.FirstName,
		LastName:    userUpdate.LastName,
		Password:    userUpdate.Password,
		NewPassword: userUpdate.NewPassword,
		Email:       userUpdate.Email,
		Avatar:      &userUpdate.Avatar,
	}

	result, err := pm.service.UserUpdate(ctx, userModelUpdate)
	if err != nil {
		return nil, err
	}

	user := result.Body.(*model.UserGet)
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
