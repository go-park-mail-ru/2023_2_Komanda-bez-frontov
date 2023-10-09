package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
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
	AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, string, error)
	AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, string, error)
	AuthLogout(ctx context.Context) (*resp.Response, string, error)
	IsSessionValid(ctx context.Context, sessionID string) (bool, error)
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

func generateSessionID(username string) string {
	s := fmt.Sprintf("%s-%d", username, time.Now().UnixMilli())
	h := sha256.New()
	h.Write([]byte(s))

	return fmt.Sprintf("%x", h.Sum(nil))
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

	hasher := md5.New()
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

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	err = s.userRepository.Insert(ctx, &repository.User{
		Username: user.Username,
		Password: encPassword,
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

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing.Password != encPassword {
		return resp.NewResponse(http.StatusUnauthorized, nil), "", fmt.Errorf("invalid username or password")
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
	currentUser := ctx.Value(model.CurrentUserInContext).(*model.UserGet)
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

func (s *authService) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
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

	currentUser, err := s.userRepository.FindByUsername(ctx, sessionInDB.Username)
	if err != nil {
		return false, err
	}

	if currentUser == nil {
		return false, nil
	}

	return true, nil
}
