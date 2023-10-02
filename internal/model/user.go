package model

type User struct {
	ID int `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type UserList struct {
	CollectionResponse
	Users []*User `json:"users"`
}
