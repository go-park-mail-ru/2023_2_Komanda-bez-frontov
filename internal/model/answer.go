package model

type Answer struct {
	ID   *int64 `json:"id"`
	Text string `json:"text" validate:"required"`
}

type AnswerResult struct {
	ID                  int64 `json:"id"`
	Text                string `json:"text"`
	SelectedTimesAnswer int    `json:"selected_times"`
}
