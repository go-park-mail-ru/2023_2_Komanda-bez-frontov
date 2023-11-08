package model

type Question struct {
	ID          *int64    `json:"id"`
	Title       string    `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        int       `json:"type" validate:"required,oneof=1 2 3"`
	Shuffle     bool      `json:"shuffle,omitempty"`
	Answers     []*Answer `json:"answers,omitempty"`
}
