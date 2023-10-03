package model

type UserLogin struct {
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,sha512"`
}

type UserSignUp struct {
	Username       string `json:"username" validate:"required,alphanum"`
	Password       string `json:"password" validate:"required,sha512"`
	PasswordRepeat string `json:"password_repeat" validate:"required,eqfield=Password"`
	Email          string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserGet struct {
	Username string `json:"username" validate:"required,alphanum"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
}

type UserList struct {
	CollectionResponse
	Users []*UserGet `json:"users" validate:"required"`
}
