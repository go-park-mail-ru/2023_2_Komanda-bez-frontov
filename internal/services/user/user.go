package user

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/serviceresponse"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Service interface {
	UserSave(ctx context.Context, user *model.User) (*resp.Response, error)
	UserList(ctx context.Context) (*resp.Response, error)
	UserGet(ctx context.Context, name string) (*resp.Response, error)
	UserVerification(ctx context.Context, user *model.User) (*resp.Response, error)
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

func (s *userService) UserSave(ctx context.Context, user *model.User) (*resp.Response, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), err
	}

	existing, err := s.userRepository.FindByName(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), nil
	}

	err = s.userRepository.Insert(ctx, s.userRepository.FromModel(user))
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, user), nil
}

func (s *userService) UserList(ctx context.Context) (*resp.Response, error) {
	var response model.UserList
	response.Users = make([]*model.User, 0)

	users, err := s.userRepository.FindAll(ctx)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	for _, user := range users {
		response.Users = append(response.Users, s.userRepository.ToModel(user))
	}

	response.Count = len(users)
	return resp.NewResponse(http.StatusOK, response), nil
}

func (s *userService) UserGet(ctx context.Context, name string) (*resp.Response, error) {
	user, err := s.userRepository.FindByName(ctx, name)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if user == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, s.userRepository.ToModel(user)), nil
}

func (s *userService) UserVerification(ctx context.Context, user *model.User) (*resp.Response, error) {
	userDB, err := s.userRepository.FindByName(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if userDB == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	if userDB.Password != user.Password {
		return resp.NewResponse(http.StatusForbidden, nil), nil
	}

	return resp.NewResponse(http.StatusOK, nil), nil
}
