package model

type CollectionResponse struct {
	Count int `json:"count"`
}

type Error struct {
	Status *string `json:"status,omitempty"`
	Code   *string `json:"code,omitempty"`
}

type ErrorResponse struct {
	Errors *[]Error `json:"errors,omitempty"`
}
