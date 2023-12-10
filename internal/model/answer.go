package model

import "github.com/microcosm-cc/bluemonday"

type Answer struct {
	ID   *int64 `json:"id"`
	Text string `json:"text" validate:"required"`
}

type AnswerResult struct {
	Text                string `json:"text"`
	SelectedTimesAnswer int    `json:"selected_times"`
}

func (answer *AnswerResult) Sanitize(sanitizer *bluemonday.Policy) {
	answer.Text = sanitizer.Sanitize(answer.Text)
}

func (answer *Answer) Sanitize(sanitizer *bluemonday.Policy) {
	answer.Text = sanitizer.Sanitize(answer.Text)
}
