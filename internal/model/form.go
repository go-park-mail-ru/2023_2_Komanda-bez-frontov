package model

type Form struct {
	Title string `json:"title" validate:"required"`
}

type FormList struct {
	CollectionResponse
	Forms []*Form `json:"forms"`
}
