package model

type Question struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        string    `json:"type" validate:"required,oneof=single_choice multiple_choice no_choice"`
	Shuffle     bool      `json:"shuffle" validate:"required"`
	Answers     []*Answer `json:"answers,omitempty"`
}
