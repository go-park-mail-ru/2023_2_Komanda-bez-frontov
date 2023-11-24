package model

const (
	SingleAnswerType   = 1
	MultipleAnswerType = 2
	InputAnswerType    = 3
)

type Question struct {
	ID          *int64    `json:"id"`
	Title       string    `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        int       `json:"type" validate:"required,oneof=1 2 3"`
	Required    bool      `json:"required"`
	Answers     []*Answer `json:"answers,omitempty"`
}
