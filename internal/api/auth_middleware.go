package api

import (
	"context"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"
	"net/http"
	"time"
)

func AuthMiddleware(sessionRepository repository.SessionRepository, userRepository repository.UserRepository, cookieExpiration time.Duration) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if err != nil {
				HandleError(w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			sessionInDB, err := sessionRepository.FindByID(r.Context(), session.Value)
			if err != nil {
				HandleError(w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if sessionInDB == nil {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				HandleError(w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			if sessionInDB.CreatedAt+cookieExpiration.Milliseconds() < time.Now().UnixMilli() {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				HandleError(w, fmt.Errorf("session expired"), &resp.Response{StatusCode: http.StatusForbidden})
				return
			}

			currentUser, err := userRepository.FindByID(r.Context(), sessionInDB.UserID)
			if err != nil {
				HandleError(w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if currentUser == nil {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				HandleError(w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), model.CurrentUserInContext, &model.UserGet{
				ID:       currentUser.ID,
				Username: currentUser.Username,
				Email:    currentUser.Email,
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
