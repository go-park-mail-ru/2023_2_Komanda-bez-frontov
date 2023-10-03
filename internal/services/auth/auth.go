package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type Service interface {
	AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, string, error)
	AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, string, error)
	AuthLogout(ctx context.Context) (*resp.Response, string, error)
}

type authService struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	validate          *validator.Validate
}

func NewAuthService(userRepository repository.UserRepository, sessionRepository repository.SessionRepository, validate *validator.Validate) Service {
	return &authService{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		validate:          validate,
	}
}

func generateSessionID(username string) string {
	s := fmt.Sprintf("%s-%d", username, time.Now().UnixMilli())
	h := sha256.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *authService) AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, string, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), "", err
	}

	existing, err := s.userRepository.FindByUsername(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), "", nil
	}

	err = s.userRepository.Insert(ctx, &repository.User{
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	sessionID := generateSessionID(user.Username)
	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: sessionID,
		Username:  user.Username,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		Username: user.Username,
		Email:    user.Email,
	}), sessionID, nil
}

func (s *authService) AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, string, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), "", err
	}

	existing, err := s.userRepository.FindByUsername(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusUnauthorized, nil), "", nil
	}

	if existing.Password != user.Password {
		return resp.NewResponse(http.StatusUnauthorized, nil), "", nil
	}

	sessionID := generateSessionID(user.Username)
	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: sessionID,
		Username:  existing.Username,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		Username: existing.Username,
		Email:    existing.Email,
	}), sessionID, nil
}

func (s *authService) AuthLogout(ctx context.Context) (*resp.Response, string, error) {
	currentUser := ctx.Value(model.CurrentUser("current_user")).(*model.UserGet)
	session, err := s.sessionRepository.FindByUsername(ctx, currentUser.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	err = s.sessionRepository.Delete(ctx, session.SessionID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusNoContent, nil), session.SessionID, nil
}
