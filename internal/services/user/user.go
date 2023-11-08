package user

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
	"net/http"

	validator "github.com/go-playground/validator/v10"
)

type Service interface {
	UserList(ctx context.Context) (*resp.Response, error)
	UserGet(ctx context.Context, id int64) (*resp.Response, error)
	UserGetAvatar(ctx context.Context, username string) (*resp.Response, error)
}

type userService struct {
	userRepository repository.UserRepository
	validate       *validator.Validate
}

func NewUserService(userRepository repository.UserRepository, validate *validator.Validate) Service {
	return &userService{
		userRepository: userRepository,
		validate:       validate,
	}
}

func (s *userService) UserList(ctx context.Context) (*resp.Response, error) {
	var response model.UserList
	response.Users = make([]*model.UserGet, 0)

	users, err := s.userRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	for _, user := range users {
		response.Users = append(response.Users, &model.UserGet{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Avatar:    user.Avatar,
		})
	}

	response.Count = len(users)
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *userService) UserGet(ctx context.Context, id int64) (*resp.Response, error) {
	user, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if user == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Avatar:    user.Avatar,
	}), nil
}

func (s *userService) UserGetAvatar(ctx context.Context, username string) (*resp.Response, error) {
	user, err := s.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if user == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, &model.UserAvatarGet{
		Username: user.Username,
		Avatar:   user.Avatar,
	}), nil
}

func (s *userService) UserUpdate(_ context.Context, _ int64, _ *model.UserSignUp) (*resp.Response, error) {
	return &resp.Response{StatusCode: http.StatusNotImplemented}, nil
}
