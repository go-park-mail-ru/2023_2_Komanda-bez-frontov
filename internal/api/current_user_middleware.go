package api

import (
	"context"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	"net/http"
)

func CurrentUserMiddleware(sessionRepository repository.SessionRepository, userRepository repository.UserRepository) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sessionInDB, err := sessionRepository.FindByID(r.Context(), session.Value)
			if err != nil || sessionInDB == nil {
				next.ServeHTTP(w, r)
				return
			}

			currentUser, err := userRepository.FindByID(r.Context(), sessionInDB.UserID)
			if err != nil || currentUser == nil {
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), model.CurrentUserInContext, &model.UserGet{
				ID:        currentUser.ID,
				Username:  currentUser.Username,
				FirstName: currentUser.FirstName,
				LastName:  currentUser.LastName,
				Email:     currentUser.Email,
				Avatar:    currentUser.Avatar,
			}))
			next.ServeHTTP(w, r)
		})
	}
}
