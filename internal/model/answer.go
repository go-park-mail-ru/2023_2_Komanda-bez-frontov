package model

type Answer struct {
	Text string `json:"text" validate:"required"`
}
