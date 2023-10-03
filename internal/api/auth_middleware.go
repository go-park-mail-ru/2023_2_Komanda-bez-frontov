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
				DefaultErrorHandler(w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if sessionDatabase == nil {
				DefaultErrorHandler(w, fmt.Errorf("session not found"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			if sessionDatabase.CreatedAt+int64(1000*60*60*24) < time.Now().UnixMilli() {
				DefaultErrorHandler(w, fmt.Errorf("session expired"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			currentUser, err := userRepository.FindByUsername(r.Context(), sessionDatabase.Username)
			if err != nil {
				DefaultErrorHandler(w, err, &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), model.CurrentUser("current_user"), currentUser))
			next.ServeHTTP(w, r)
		})
	}
}
