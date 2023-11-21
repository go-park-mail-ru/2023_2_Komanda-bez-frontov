package model

type Answer struct {
	ID   *int64 `json:"id"`
	Text string `json:"text" validate:"required"`
}

type AnswerResult struct {
	Description     string `json:"text"`
	SelectedTimes   int    `json:"selected_times"`
	NumberOfPassages int   `json:"number_of_passages"`
}