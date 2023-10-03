package model

type User struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type UserList struct {
	CollectionResponse
	Users []*User `json:"users"`
}
