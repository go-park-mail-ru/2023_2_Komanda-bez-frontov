package model

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
	ID        int64   `json:"id" validate:"required,uuid"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Username  string  `json:"username" validate:"required,alphanum"`
	Email     string  `json:"email,omitempty" validate:"omitempty,email"`
	Avatar    *string `json:"avatar,omitempty"`
}

type UserList struct {
	CollectionResponse
	Users []*UserGet `json:"users" validate:"required"`
}
