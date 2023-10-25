package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // nolint:gosec
	"encoding/hex"
	"fmt"
	"go-form-hub/internal/config"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
	"io"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
)

type Service interface {
	AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, int64, error)
	AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, int64, error)
	AuthLogout(ctx context.Context) (*resp.Response, int64, error)
	IsSessionValid(ctx context.Context, sessionID int64) (bool, error)
}

type authService struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	cfg               *config.Config
	validate          *validator.Validate
}

func NewAuthService(userRepository repository.UserRepository, sessionRepository repository.SessionRepository, cfg *config.Config, validate *validator.Validate) Service {
	return &authService{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		cfg:               cfg,
		validate:          validate,
	}
}

func (s *authService) encryptPassword(pass string) (string, error) {
	keyBytes, err := hex.DecodeString(s.cfg.EncryptionKey)
	if err != nil {
		return "", fmt.Errorf("encrypt_password invalid hex-encoded key: %e", err)
	}

	if len(keyBytes) != 32 {
		return "", fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	hasher := md5.New() // nolint:gosec
	_, err = io.WriteString(hasher, pass)
	if err != nil {
		return "", err
	}
	nonce := hasher.Sum(nil)[:12]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(pass), nil)

	return hex.EncodeToString(nonce) + hex.EncodeToString(ciphertext), nil
}

func (s *authService) AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, int64, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), 0, err
	}

	existing, err := s.userRepository.FindByUsername(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), 0, nil
	}

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	id, err := s.userRepository.Insert(ctx, &repository.User{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  encPassword,
		Email:     user.Email,
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: 1,
		UserID:    id,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		ID:        id,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}), 0, nil
}

func (s *authService) AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, int64, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), 0, err
	}

	existing, err := s.userRepository.FindByUsername(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusUnauthorized, nil), 0, nil
	}

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	if existing.Password != encPassword {
		return resp.NewResponse(http.StatusUnauthorized, nil), 0, fmt.Errorf("invalid username or password")
	}

	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: 1,
		UserID:    existing.ID,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		ID:        existing.ID,
		FirstName: existing.FirstName,
		LastName:  existing.LastName,
		Username:  existing.Username,
		Email:     existing.Email,
	}), 1, nil
}

func (s *authService) AuthLogout(ctx context.Context) (*resp.Response, int64, error) {
	currentUser := ctx.Value(model.CurrentUserInContext).(*model.UserGet)
	session, err := s.sessionRepository.FindByUserID(ctx, currentUser.ID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	err = s.sessionRepository.Delete(ctx, session.SessionID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), 0, err
	}

	return resp.NewResponse(http.StatusNoContent, nil), session.SessionID, nil
}

func (s *authService) IsSessionValid(ctx context.Context, sessionID int64) (bool, error) {
	sessionInDB, err := s.sessionRepository.FindByID(ctx, sessionID)
	if err != nil {
		return false, err
	}

	if sessionInDB == nil {
		return false, nil
	}

	if sessionInDB.CreatedAt+s.cfg.CookieExpiration.Milliseconds() < time.Now().UnixMilli() {
		return false, nil
	}

	currentUser, err := s.userRepository.FindByID(ctx, sessionInDB.UserID)
	if err != nil {
		return false, err
	}

	if currentUser == nil {
		return false, nil
	}

	return true, nil
}
