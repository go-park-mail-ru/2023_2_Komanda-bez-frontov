package model

import "github.com/microcosm-cc/bluemonday"

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

type QuestionResult struct {
	ID                       int64           `json:"id"`
	Title                    string          `json:"title"`
	Description              string          `json:"description"`
	Type                     int             `json:"type"`
	Required                 bool            `json:"required"`
	NumberOfPassagesQuestion int             `json:"number_of_passages"`
	Answers                  []*AnswerResult `json:"answers"`
}

func (question *Question) Sanitize(sanitizer *bluemonday.Policy) {
	question.Title = sanitizer.Sanitize(question.Title)
	if question.Description != nil {
		*question.Description = sanitizer.Sanitize(*question.Description)
	}
	for _, answer := range question.Answers {
		answer.Sanitize(sanitizer)
	}
}

func (question *QuestionResult) Sanitize(sanitizer *bluemonday.Policy) {
	question.Title = sanitizer.Sanitize(question.Title)
	question.Description = sanitizer.Sanitize(question.Description)
	for _, answer := range question.Answers {
		answer.Sanitize(sanitizer)
	}
}
