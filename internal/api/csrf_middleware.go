package api

import (
	"fmt"
	"net/http"

	resp "go-form-hub/internal/services/service_response"
)

const csrfCookieName = "csrf_token"

func CSRFMiddleware(tokenParser *HashToken, responseEncoder ResponseEncoder) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			flag := false
			csrfMethods := []string{http.MethodDelete, http.MethodPost, http.MethodPatch, http.MethodPut}
			for _, method := range csrfMethods {
				if r.Method == method {
					flag = true
					break
				}
			}
			if !flag {
				next.ServeHTTP(w, r)
				return
			}

			session, err := r.Cookie(sessionCookieName)
			if err != nil {
				responseEncoder.HandleError(ctx, w, fmt.Errorf("you have to log in or sign up to continue"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			csrfToken := r.Header.Get("X-CSRF-Token")
			valid, err := tokenParser.Check(session.Value, csrfToken)
			if err != nil {
				responseEncoder.HandleError(ctx, w, fmt.Errorf("CSRF error"), &resp.Response{StatusCode: http.StatusInternalServerError})
				return
			}

			if !valid {
				responseEncoder.HandleError(ctx, w, fmt.Errorf("CSRF not passed"), &resp.Response{StatusCode: http.StatusUnauthorized})
				return
			}

			w.Header().Add("X-CSRF-Token", csrfToken)
			next.ServeHTTP(w, r)
		})
	}
}
