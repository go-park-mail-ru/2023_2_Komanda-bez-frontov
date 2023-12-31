package controller

import (
	"context"

	"go-form-hub/internal/model"
	"go-form-hub/microservices/auth/session"
	"go-form-hub/microservices/auth/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthController struct {
	session.UnimplementedAuthCheckerServer

	authUseCase usecase.AuthUseCase
	validator   *validator.Validate
}

func NewAuthController(authUsecase usecase.AuthUseCase, v *validator.Validate) *AuthController {
	return &AuthController{
		authUseCase: authUsecase,
		validator:   v,
	}
}

func (m *AuthController) Login(ctx context.Context, userLogin *session.UserLogin) (*session.SessionInfo, error) {
	user := model.UserLogin{
		Email:    userLogin.Email,
		Password: userLogin.Password,
	}

	response, sessionID, err := m.authUseCase.AuthLogin(ctx, &user)
	if err != nil {
		log.Error().Msgf("error logging in: %v", err)
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	userInfo := response.Body.(*model.UserGet)
	if userInfo.Avatar == nil {
		userInfo.Avatar = new(string)
	}

	userMsg := &session.User{
		Id:        userInfo.ID,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Avatar:    *userInfo.Avatar,
	}

	res := &session.SessionInfo{
		Session:     sessionID,
		CurrentUser: userMsg,
	}

	return res, nil
}

func (m *AuthController) Signup(ctx context.Context, userSignup *session.UserSignup) (*session.SessionInfo, error) {
	user := model.UserSignUp{
		Email:     userSignup.Email,
		Password:  userSignup.Password,
		FirstName: userSignup.FirstName,
		LastName:  userSignup.LastName,
		Username:  userSignup.Username,
	}

	response, sessionID, err := m.authUseCase.AuthSignUp(ctx, &user)
	if err != nil {
		log.Error().Msgf("error signing up: %v", err)
		return nil, status.Errorf(codes.Canceled, err.Error())
	}

	userInfo := response.Body.(*model.UserGet)
	if userInfo.Avatar == nil {
		userInfo.Avatar = new(string)
	}

	userMsg := &session.User{
		Id:        userInfo.ID,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Avatar:    *userInfo.Avatar,
	}

	res := &session.SessionInfo{
		Session:     sessionID,
		CurrentUser: userMsg,
	}

	return res, nil
}

func (m *AuthController) Check(ctx context.Context, sessionID *session.Session) (*session.CheckResult, error) {
	valid, err := m.authUseCase.IsSessionValid(ctx, sessionID.Session)
	if err != nil {
		log.Error().Msgf("error finding session: %v", err)
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	res := &session.CheckResult{
		Valid: valid,
	}

	return res, nil
}

func (m *AuthController) Delete(ctx context.Context, sessionID *session.Session) (*session.Nothing, error) {
	_, _, err := m.authUseCase.AuthLogout(ctx, sessionID.Session)
	if err != nil {
		log.Error().Msgf("error logging in: %v", err)
		return nil, err
	}
	return &session.Nothing{}, nil
}
