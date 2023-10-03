package api

import (
	"encoding/json"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	"go-form-hub/internal/services/user"
	"io"
	"time"
	"net/http"
	
	"github.com/go-playground/validator/v10"
	// jwt "github.com/dgrijalva/jwt-go"
)

type UserAPIController struct {
	sessions  repository.SessionRepository
	service   user.Service
	validator *validator.Validate
}

func NewUserAPIController(sessions repository.SessionRepository, service user.Service, v *validator.Validate) Router {
	return &UserAPIController{
		sessions: sessions,
		service: service,
		validator: v,
	}
}

func (c *UserAPIController) Routes() []Route {
	return []Route{
		{
			Name:         "Login",
			Method:       http.MethodPost,
			Path:         "/login",
			Handler:      c.Login,
			AuthRequired: false,
		},
		{
			Name:         "Signup",
			Method:       http.MethodPost,
			Path:         "/signup",
			Handler:      c.Signup,
			AuthRequired: false,
		},
		{
			Name:         "Logout",
			Method:       http.MethodPost,
			Path:         "/logout",
			Handler:      c.Logout,
			AuthRequired: true,
		},
	}
}

func (c *UserAPIController) Login(w http.ResponseWriter, r *http.Request) {

	session, err := r.Cookie("session_id")
	if err != http.ErrNoCookie {
		if c.sessions.FindByID(r.Context(), session.Value) {
			http.Error(w, `Previous session is not terminated`, 403)
			return
		}
	}
	
	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	var user model.User
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.UserVerification(r.Context(), &user)
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}
	
	if result.StatusCode == http.StatusNotFound {
		http.Error(w, `Not found`, 404)
		return
	}

	if result.StatusCode == http.StatusForbidden {
		http.Error(w, `Bad password`, 400)
		return
	}

	SID, err := c.sessions.Insert(r.Context(), user.Username)
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}
	http.SetCookie(w, cookie)
	w.Write([]byte(SID + string('\n')))

	EncodeJSONResponse(`Login Success`, http.StatusOK, w)
}

func (c *UserAPIController) Signup(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	var user model.User
	if err = json.Unmarshal(requestJSON, &user); err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	result, err := c.service.UserSave(r.Context(), &user)
	if err != nil {
		EncodeJSONResponse(err, result.StatusCode, w)
		return
	}

	if result.StatusCode == 409 {
		EncodeJSONResponse(`User exists`, result.StatusCode, w)
		return
	}

	EncodeJSONResponse(`Signup Success`, result.StatusCode, w)
}

func (c *UserAPIController) Logout(w http.ResponseWriter, r *http.Request) {

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Error(w, `No cookie`, 401)
		return
	}

	if !c.sessions.FindByID(r.Context(), session.Value) {
		http.Error(w, `No session`, 401)
		return
	}

	err = c.sessions.Delete(r.Context(), session.Value)
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}

	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)

	EncodeJSONResponse(`Logout Success`, http.StatusOK, w)
}

