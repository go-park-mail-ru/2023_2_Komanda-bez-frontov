package model

import "time"

type Form struct {
	ID        *int64      `json:"id"`
	Title     string      `json:"title" validate:"required"`
	Anonymous bool        `json:"anonymous"`
	Author    *UserGet    `json:"author"`
	CreatedAt time.Time   `json:"created_at"`
	Questions []*Question `json:"questions" validate:"required"`
}

type FormTitle struct {
	ID        int64     `json:"id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

type FormList struct {
	CollectionResponse
	Forms []*Form `json:"forms" validate:"required"`
}

type FormTitleList struct {
	CollectionResponse
	FormTitles []*FormTitle `json:"forms" validate:"required"`
}

type FormPassage struct {
	FormID         *int64           `json:"form_id" validate:"required"`
	PassageAnswers []*PassageAnswer `json:"passage_answers" validate:"required"`
}

type PassageAnswer struct {
	QuestionID *int64 `json:"question_id" validate:"required"`
	Text       string `json:"answer_text"`
}
