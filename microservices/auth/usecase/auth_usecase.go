package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // nolint:gosec
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-form-hub/internal/config"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
)

var (
	ErrUsernameTaken    = errors.New("username taken")
	ErrEmailTaken       = errors.New("email taken")
	ErrWrongCredentials = errors.New("login credentials are wrong")
)

type AuthUseCase interface {
	AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, string, error)
	AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, string, error)
	AuthLogout(ctx context.Context, sessionID string) (*resp.Response, string, error)
	IsSessionValid(ctx context.Context, sessionID string) (bool, error)
}

type authUseCase struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	cfg               *config.Config
	validate          *validator.Validate
}

func NewAuthUseCase(userRepository repository.UserRepository, sessionRepository repository.SessionRepository, cfg *config.Config, validate *validator.Validate) AuthUseCase {
	return &authUseCase{
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

func (s *authUseCase) encryptPassword(pass string) (string, error) {
	keyBytes, err := hex.DecodeString(s.cfg.EncryptionKey)
	if err != nil {
		return "", fmt.Errorf("encrypt_password invalid hex-encoded key: %v", err)
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

func (s *authUseCase) AuthSignUp(ctx context.Context, user *model.UserSignUp) (*resp.Response, string, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), "", err
	}

	existing, err := s.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), "", ErrEmailTaken
	}

	existing, err = s.userRepository.FindByUsername(ctx, user.Username)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing != nil {
		return resp.NewResponse(http.StatusConflict, nil), "", ErrUsernameTaken
	}

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	id, err := s.userRepository.Insert(ctx, &repository.User{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  encPassword,
		Email:     user.Email,
		Avatar:    user.Avatar,
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	sessionID := generateSessionID(user.Username)
	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: sessionID,
		UserID:    id,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		ID:        id,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}), sessionID, nil
}

func (s *authUseCase) AuthLogin(ctx context.Context, user *model.UserLogin) (*resp.Response, string, error) {
	if err := s.validate.Struct(user); err != nil {
		return resp.NewResponse(http.StatusBadRequest, nil), "", err
	}

	existing, err := s.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing == nil {
		return resp.NewResponse(http.StatusUnauthorized, nil), "", ErrWrongCredentials
	}

	encPassword, err := s.encryptPassword(user.Password)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	if existing.Password != encPassword {
		return resp.NewResponse(http.StatusUnauthorized, nil), "", fmt.Errorf("invalid username or password")
	}

	sessionID := generateSessionID(existing.Username)
	err = s.sessionRepository.Insert(ctx, &repository.Session{
		SessionID: sessionID,
		UserID:    existing.ID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusOK, &model.UserGet{
		ID:        existing.ID,
		FirstName: existing.FirstName,
		LastName:  existing.LastName,
		Username:  existing.Username,
		Email:     existing.Email,
	}), sessionID, nil
}

func (s *authUseCase) AuthLogout(ctx context.Context, sessionID string) (*resp.Response, string, error) {
	err := s.sessionRepository.Delete(ctx, sessionID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), "", err
	}

	return resp.NewResponse(http.StatusNoContent, nil), sessionID, nil
}

func (s *authUseCase) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
	sessionInDB, err := s.sessionRepository.FindByID(ctx, sessionID)
	if err != nil {
		return false, err
	}

	if sessionInDB == nil {
		return false, nil
	}

	if sessionInDB.CreatedAt.UnixMilli()+s.cfg.CookieExpiration.Milliseconds() < time.Now().UTC().UnixMilli() {
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
