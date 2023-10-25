package model

type CurrentUserInContextType string

const CurrentUserInContext = CurrentUserInContextType("current_user")

type UserLogin struct {
	Username  string `json:"username" validate:"required,alphanum"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Password  string `json:"password" validate:"required"`
}

type UserSignUp struct {
	Username       string `json:"username" validate:"required,alphanum"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	Password       string `json:"password" validate:"required,sha512"`
	PasswordRepeat string `json:"password_repeat" validate:"required,eqfield=Password"`
	Email          string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserGet struct {
	ID        int64  `json:"id" validate:"required,uuid"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username" validate:"required,alphanum"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserList struct {
	CollectionResponse
	Users []*UserGet `json:"users" validate:"required"`
}
