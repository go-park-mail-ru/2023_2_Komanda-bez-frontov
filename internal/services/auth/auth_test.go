package auth_test

import (
	"context"
	"go-form-hub/internal/config"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	mocks "go-form-hub/internal/repository/mocks"
	"go-form-hub/internal/services/auth"
	"net/http"
	"testing"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var (
	ctx               = context.Background()
	validate          = validator.New()
	userRepository    = mocks.NewUserMockRepository()
	sessionRepository = mocks.NewSessionMockRepository()
	cfg               = &config.Config{CookieExpiration: 1 * time.Nanosecond, EncryptionKey: "1248712441dbbf43bb37f91d626a020e7e0f4486f050142034b8a267b06a2f0c"}
	authService       = auth.NewAuthService(userRepository, sessionRepository, cfg, validate)
	sha512String      = "1f40fc92da241694750979ee6cf582f2d5d7d28e18335de05abc54d0560e0f5302860c652bf08d560252aa5e74210546f369fbbbce8c12cfc7957b2652fe9a75"
	sha512String2     = "5267768822ee624d48fce15ec5ca79cbd602cb7f4c2157a516556991f22ef8c7b5ef7b18d1ff41c59370efb0858651d44a936c11b7b144c48fe04df3c6a3e8da"
)

func TestAuthSignUp(t *testing.T) {
	t.Run("UsernameIsRequired", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Name: "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'Username' failed on the 'required' tag")
	})

	t.Run("PasswordIsRequired", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Name: "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'Password' failed on the 'required' tag")
	})

	t.Run("PasswordRepeatIsRequired", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Name: "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'PasswordRepeat' failed on the 'required' tag")
	})

	t.Run("PasswordShouldBeEqualToRepeat", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "test",
			Password:       sha512String,
			PasswordRepeat: sha512String2,
			Name:           "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'PasswordRepeat' failed on the 'eqfield' tag")
	})

	t.Run("PasswordShoulBeInSHA512", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "test",
			Password:       "avd",
			PasswordRepeat: "avd",
			Name:           "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'Password' failed on the 'sha512' tag")
	})

	t.Run("EmailShouldBeInRightFormat", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "test",
			Password:       "avd",
			PasswordRepeat: "avd",
			Name:           "test",
			Email:          "invalid-email",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'Email' failed on the 'email' tag")
	})

	t.Run("UsernameShouldBeAlphanumeric", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "__!!@#@!$",
			Password:       "avd",
			PasswordRepeat: "avd",
			Name:           "test",
			Email:          "test",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.NotNil(t, err) || !assert.Equal(t, http.StatusBadRequest, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "Field validation for 'Username' failed on the 'alphanum' tag")
	})

	t.Run("UsernameShouldBeUnique", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "ValidUsername",
			Password:       sha512String,
			PasswordRepeat: sha512String,
			Name:           "test",
			Email:          "test@test.com",
		}

		err := userRepository.Insert(ctx, &repository.User{
			Username: userSignUp.Username,
			Name:     userSignUp.Name,
			Email:    userSignUp.Email,
			Surname:  userSignUp.Surname,
			Password: userSignUp.Password,
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into user repository: %e", err)
			t.FailNow()
		}

		t.Cleanup(func() {
			_ = userRepository.Delete(ctx, userSignUp.Username)
		})

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.Nil(t, err) || !assert.Equal(t, http.StatusConflict, r.StatusCode) || !assert.Equal(t, "", sessionID) {
			t.Logf("err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}
	})

	t.Run("ValidUser", func(t *testing.T) {
		t.Parallel()
		userSignUp := model.UserSignUp{
			Username:       "ValidUniqueUsername",
			Password:       sha512String,
			PasswordRepeat: sha512String,
			Name:           "test",
			Email:          "test@test.com",
		}

		r, sessionID, err := authService.AuthSignUp(ctx, &userSignUp)
		if !assert.Nil(t, err) || !assert.Equal(t, http.StatusOK, r.StatusCode) || !assert.NotEqual(t, "", sessionID) {
			t.Logf("failed to signup: err: %e, sessionID: %s, status: %d", err, sessionID, r.StatusCode)
			t.FailNow()
		}
		t.Cleanup(func() {
			_ = userRepository.Delete(ctx, userSignUp.Username)
			_ = sessionRepository.Delete(ctx, sessionID)
		})

		user, err := userRepository.FindByUsername(ctx, userSignUp.Username)
		if !assert.Nil(t, err) || !assert.NotNil(t, user) {
			t.Logf("failed to find user in user repository: %e", err)
			t.FailNow()
		}

		assert.NotEqual(t, "", user.ID)

		session, err := sessionRepository.FindByID(ctx, sessionID)
		if !assert.Nil(t, err) || !assert.NotNil(t, session) {
			t.Logf("failed to find session in session repository: %e", err)
			t.FailNow()
		}

		assert.Equal(t, user.ID, session.UserID)
	})
}
