package model

import "time"

const AnonUserID = 0

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

type FormResult struct {
	ID                   int64             `json:"id"`
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	CreatedAt            time.Time         `json:"created_at"`
	Author               *UserGet          `json:"author"`
	NumberOfPassagesForm int               `json:"number_of_passages"`
	Questions            []*QuestionResult `json:"questions"`
	Anonymous            bool              `json:"anonymous"`
	Participants         []*UserGet        `json:"participants,omitempty"`
}

type FormPassage struct {
	FormID         *int64           `json:"form_id" validate:"required"`
	PassageAnswers []*PassageAnswer `json:"passage_answers" validate:"required"`
}

type PassageAnswer struct {
	QuestionID *int64 `json:"question_id" validate:"required"`
	Text       string `json:"answer_text"`
}

type FormPassageResult struct {
	FormID     int64  `json:"form_id"`
	UserID     int64  `json:"user_id" db:"user_id"`
	Username   string `json:"username" db:"username"`
	FirstName  string `json:"first_name" db:"first_name"`
	LastName   string `json:"last_name" db:"last_name"`
	Email      string `json:"email" db:"email"`
	QuestionID int64  `json:"question_id" db:"question_id"`
	AnswerText string `json:"answer_text" db:"answer_text"`
}
