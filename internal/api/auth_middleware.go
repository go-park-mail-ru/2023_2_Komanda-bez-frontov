package api

import (
	"context"
	"go-form-hub/internal/repository"
	"net/http"
)

func AuthMiddleware(sessionRepository repository.SessionRepository, userRepository repository.UserRepository) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sessionDatabase, err := sessionRepository.FindByID(r.Context(), session.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if sessionDatabase == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			currentUser, err := userRepository.FindByUsername(r.Context(), sessionDatabase.Username)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), "current_user", currentUser))
			next.ServeHTTP(w, r)
		})
	}
}
