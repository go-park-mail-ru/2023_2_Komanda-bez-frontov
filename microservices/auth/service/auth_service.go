package session_service

import (
	"context"

	"go-form-hub/internal/model"
	"go-form-hub/internal/services/auth"
	"go-form-hub/microservices/auth/session"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthManager struct {
	session.UnimplementedAuthCheckerServer

	authService auth.Service
	validator   *validator.Validate
}

func NewAuthManager(authService auth.Service, v *validator.Validate) *AuthManager {
	return &AuthManager{
		authService: authService,
		validator:   v,
	}
}

func (m *AuthManager) Login(ctx context.Context, userLogin *session.UserLogin) (*session.Session, error) {
	user := model.UserLogin{
		Email:    userLogin.Email,
		Password: userLogin.Password,
	}

	_, sessionID, err := m.authService.AuthLogin(ctx, &user)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	res := &session.Session{
		Session: sessionID,
	}

	return res, nil
}

func (m *AuthManager) Signup(ctx context.Context, userSignup *session.UserSignup) (*session.Session, error) {
	user := model.UserSignUp{
		Email:     userSignup.Email,
		Password:  userSignup.Password,
		FirstName: userSignup.FirstName,
		LastName:  userSignup.LastName,
		Username:  userSignup.Username,
	}

	_, sessionID, err := m.authService.AuthSignUp(ctx, &user)
	if err != nil {
		return nil, status.Errorf(codes.Canceled, err.Error())
	}

	res := &session.Session{
		Session: sessionID,
	}

	return res, nil
}

func (m *AuthManager) Check(ctx context.Context, sessionID *session.Session) (*session.CheckResult, error) {
	valid, err := m.authService.IsSessionValid(ctx, sessionID.Session)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	res := &session.CheckResult{
		Valid: valid,
	}

	return res, nil
}

func (m *AuthManager) Delete(ctx context.Context, sessionID *session.Session) (*session.Nothing, error) {
	m.authService.AuthLogout(ctx)

	return &session.Nothing{}, nil
}
