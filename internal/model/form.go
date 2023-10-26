package model

import "time"

type Form struct {
	ID        *int64    `json:"id"`
	Title     string    `json:"title" validate:"required"`
	AuthorID  int64     `json:"author_id" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type FormList struct {
	CollectionResponse
	Forms []*Form `json:"forms" validate:"required"`
}
