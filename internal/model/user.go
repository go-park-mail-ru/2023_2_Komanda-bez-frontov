package model

import (
	"github.com/microcosm-cc/bluemonday"
)

type ContextCurrentUserType string

const ContextCurrentUser = ContextCurrentUserType("current_user")

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserSignUp struct {
	Username  string  `json:"username" validate:"required,alphanum"` // мб удалить
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Password  string  `json:"password" validate:"required"`
	Email     string  `json:"email" validate:"required,email"`
	Avatar    *string `json:"avatar,omitempty"`
}

type UserGet struct {
	ID        int64   `json:"id" validate:"required"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Username  string  `json:"username" validate:"required,alphanum"`
	Email     string  `json:"email,omitempty" validate:"omitempty,email"`
	Avatar    *string `json:"avatar,omitempty"`
}

func (user *UserGet) Sanitize(sanitizer *bluemonday.Policy) {
	user.Username = sanitizer.Sanitize(user.Username)
	user.FirstName = sanitizer.Sanitize(user.FirstName)
	user.LastName = sanitizer.Sanitize(user.LastName)
	user.Email = sanitizer.Sanitize(user.Email)
}

type UserUpdate struct {
	Username    string  `json:"username" validate:"required,alphanum"`
	FirstName   string  `json:"first_name,omitempty"`
	LastName    string  `json:"last_name,omitempty"`
	Password    string  `json:"oldPassword,omitempty"`
	NewPassword string  `json:"newPassword,omitempty"`
	Email       string  `json:"email" validate:"required,email"`
	Avatar      *string `json:"avatar,omitempty"`
}

type UserAvatarGet struct {
	Username string  `json:"username" validate:"required,alphanum"`
	Avatar   *string `json:"avatar" validate:"required"`
}

func (user *UserAvatarGet) Sanitize(sanitizer *bluemonday.Policy) {
	user.Username = sanitizer.Sanitize(user.Username)
}

type UserList struct {
	CollectionResponse
	Users []*UserGet `json:"users" validate:"required"`
}

func (users *UserList) Sanitize(sanitizer *bluemonday.Policy) {
	for _, user := range users.Users {
		user.Sanitize(sanitizer)
	}
}
