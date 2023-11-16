package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
)

func AuthMiddleware(sessionRepository repository.SessionRepository, userRepository repository.UserRepository, cookieExpiration time.Duration, responseEncoder ResponseEncoder) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, err := r.Cookie("session_id")
			if err != nil {
				responseEncoder.HandleError(ctx, w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			sessionInDB, err := sessionRepository.FindByID(r.Context(), session.Value)
			if err != nil {
				responseEncoder.HandleError(ctx, w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if sessionInDB == nil {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				responseEncoder.HandleError(ctx, w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			if sessionInDB.CreatedAt.UnixMilli()+cookieExpiration.Milliseconds() < time.Now().UTC().UnixMilli() {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				responseEncoder.HandleError(ctx, w, fmt.Errorf("session expired"), &resp.Response{StatusCode: http.StatusForbidden})
				return
			}

			currentUser, err := userRepository.FindByID(r.Context(), sessionInDB.UserID)
			if err != nil {
				responseEncoder.HandleError(ctx, w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if currentUser == nil {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				responseEncoder.HandleError(ctx, w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), model.ContextCurrentUser, &model.UserGet{
				ID:        currentUser.ID,
				Username:  currentUser.Username,
				FirstName: currentUser.FirstName,
				LastName:  currentUser.LastName,
				Email:     currentUser.Email,
			}))
			next.ServeHTTP(w, r)
		})
	}
}

func createExpiredCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:    name,
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}
}
