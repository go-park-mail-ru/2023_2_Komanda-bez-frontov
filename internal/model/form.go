package model

import "time"

type Form struct {
	ID        *int64    `json:"id"`
	Title     string    `json:"title" validate:"required"`
	Author    *UserGet  `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type FormList struct {
	CollectionResponse
	Forms []*Form `json:"forms" validate:"required"`
}
