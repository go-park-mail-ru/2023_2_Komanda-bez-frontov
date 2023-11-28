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

	authService usecase.AuthUseCase
	validator   *validator.Validate
}

func NewAuthController(authUsecase usecase.AuthUseCase, v *validator.Validate) *AuthController {
	return &AuthController{
		authService: authUsecase,
		validator:   v,
	}
}

func (m *AuthController) Login(ctx context.Context, userLogin *session.UserLogin) (*session.Session, error) {
	user := model.UserLogin{
		Email:    userLogin.Email,
		Password: userLogin.Password,
	}

	_, sessionID, err := m.authService.AuthLogin(ctx, &user)
	if err != nil {
		log.Error().Msgf("error logging in: %v", err)
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	res := &session.Session{
		Session: sessionID,
	}

	return res, nil
}

func (m *AuthController) Signup(ctx context.Context, userSignup *session.UserSignup) (*session.Session, error) {
	user := model.UserSignUp{
		Email:     userSignup.Email,
		Password:  userSignup.Password,
		FirstName: userSignup.FirstName,
		LastName:  userSignup.LastName,
		Username:  userSignup.Username,
	}

	_, sessionID, err := m.authService.AuthSignUp(ctx, &user)
	if err != nil {
		log.Error().Msgf("error signing up: %v", err)
		return nil, status.Errorf(codes.Canceled, err.Error())
	}

	res := &session.Session{
		Session: sessionID,
	}

	return res, nil
}

func (m *AuthController) Check(ctx context.Context, sessionID *session.Session) (*session.CheckResult, error) {
	valid, err := m.authService.IsSessionValid(ctx, sessionID.Session)
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
	_, _, err := m.authService.AuthLogout(ctx, sessionID.Session)
	if err != nil {
		log.Error().Msgf("error logging in: %v", err)
		return nil, err
	}
	return &session.Nothing{}, nil
}
