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
	UserGet(ctx context.Context, username string) (*resp.Response, error)
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
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		})
	}

	response.Count = len(users)
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *userService) UserGet(ctx context.Context, name string) (*resp.Response, error) {
	user, err := s.userRepository.FindByUsername(ctx, name)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if user == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}), nil
}
