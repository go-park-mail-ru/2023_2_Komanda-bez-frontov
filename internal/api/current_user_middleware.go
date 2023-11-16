package api

import (
	"context"
	"net/http"
	"time"

	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
)

func CurrentUserMiddleware(sessionRepository repository.SessionRepository, userRepository repository.UserRepository, cookieExpiration time.Duration) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sessionInDB, err := sessionRepository.FindByID(r.Context(), session.Value)
			if err != nil || sessionInDB == nil {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				next.ServeHTTP(w, r)
				return
			}

			if sessionInDB.CreatedAt.UnixMilli()+cookieExpiration.Milliseconds() < time.Now().UTC().UnixMilli() {
				cookie := createExpiredCookie("session_id")
				http.SetCookie(w, cookie)
				next.ServeHTTP(w, r)
				return
			}

			currentUser, err := userRepository.FindByID(r.Context(), sessionInDB.UserID)
			if err != nil || currentUser == nil {
				next.ServeHTTP(w, r)
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
