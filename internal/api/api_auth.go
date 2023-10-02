package api

import (
	"math/rand"
	"encoding/json"
	"go-form-hub/internal/model"
	"go-form-hub/internal/services/user"
	"io"
	"time"
	"net/http"
	
	"github.com/go-playground/validator/v10"
)

type UserAPIController struct {
	sessions map[string]uint
	service   user.Service
	validator *validator.Validate
}

func NewUserAPIController(service user.Service, v *validator.Validate) Router {
	return &UserAPIController{
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

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (c *UserAPIController) Login(w http.ResponseWriter, r *http.Request) {
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

	result, err := c.service.GetUser(r.Context(), user.Name)
	if err != nil {
		EncodeJSONResponse(err, http.StatusInternalServerError, w)
		return
	}
	
	if result == nil {
		http.Error(w, `Not found`, 404)
		return
	}

	if user.Password != result.Password {
		http.Error(w, `Bad request`, 400)
		return
	}

	SID := RandStringRunes(32)

	c.sessions[SID] = user.ID

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}
	http.SetCookie(w, cookie)
	w.Write([]byte(SID))

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

	// SID := RandStringRunes(32)

	// c.sessions[SID] = user.ID

	// cookie := &http.Cookie{
	// 	Name:    "session_id",
	// 	Value:   SID,
	// 	Expires: time.Now().Add(10 * time.Hour),
	// }
	// http.SetCookie(w, cookie)
	// //w.Write([]byte(SID))

	EncodeJSONResponse(result.Body, result.StatusCode, w)
}

func (c *UserAPIController) Logout(w http.ResponseWriter, r *http.Request) {

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Error(w, `No session`, 401)
		return
	}

	if _, ok := c.sessions[session.Value]; !ok {
		http.Error(w, `No session`, 401)
		return
	}

	delete(c.sessions, session.Value)

	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)
}

