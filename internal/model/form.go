package model

import (
	"database/sql"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const AnonUserID = 0

type Form struct {
	ID                  *int64      `json:"id"`
	Title               string      `json:"title" validate:"required"`
	Description         *string     `json:"description"`
	Anonymous           bool        `json:"anonymous"`
	PassageMax          int         `json:"passage_max"`
	CurrentPassageTotal int         `json:"cur_passage_total"`
	IsArchived          bool        `json:"is_archived"`
	ArchiveWhen         *time.Time  `json:"archive_when"`
	Author              *UserGet    `json:"author"`
	CreatedAt           time.Time   `json:"created_at"`
	Questions           []*Question `json:"questions" validate:"required"`
}

func (form *Form) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
	if form.Description != nil {
		*form.Description = sanitizer.Sanitize(*form.Description)
	}
	form.Author.Sanitize(sanitizer)
	for _, question := range form.Questions {
		question.Sanitize(sanitizer)
	}
}

type FormTitle struct {
	ID                   int64     `json:"id" validate:"required" db:"id"`
	Title                string    `json:"title" validate:"required" db:"title"`
	CreatedAt            time.Time `json:"created_at" validate:"required" db:"created_at"`
	NumberOfPassagesForm int       `json:"number_of_passages" db:"number_of_passages"`
}

func (form *FormTitle) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
}

type FormList struct {
	CollectionResponse
	Forms []*FormTitle `json:"forms" validate:"required"`
}

func (forms *FormList) Sanitize(sanitizer *bluemonday.Policy) {
	for _, form := range forms.Forms {
		form.Sanitize(sanitizer)
	}
}

type FormTitleList struct {
	CollectionResponse
	FormTitles []*FormTitle `json:"forms" validate:"required"`
}

func (forms *FormTitleList) Sanitize(sanitizer *bluemonday.Policy) {
	for _, form := range forms.FormTitles {
		form.Sanitize(sanitizer)
	}
}

type FormUpdate struct {
	ID               *int64      `json:"id"`
	Title            string      `json:"title" validate:"required"`
	Description      *string     `json:"description"`
	Anonymous        bool        `json:"anonymous"`
	PassageMax       int         `json:"passage_max"`
	IsArchived       bool        `json:"is_archived"`
	ArchiveWhen      *time.Time  `json:"archive_when"`
	Author           *UserGet    `json:"author"`
	CreatedAt        time.Time   `json:"created_at"`
	Questions        []*Question `json:"questions" validate:"required"`
	RemovedQuestions []int64     `json:"removed_questions"`
	RemovedAnswers   []int64     `json:"removed_answers"`
}

func (form *FormUpdate) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
	if form.Description != nil {
		*form.Description = sanitizer.Sanitize(*form.Description)
	}
	form.Author.Sanitize(sanitizer)
	for _, question := range form.Questions {
		question.Sanitize(sanitizer)
	}
}

type FormResult struct {
	ID                   int64             `json:"id"`
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	CreatedAt            time.Time         `json:"created_at"`
	IsArchived           bool              `json:"is_archived"`
	Author               *UserGet          `json:"author"`
	PassageMax           int               `json:"passage_max"`
	NumberOfPassagesForm int               `json:"number_of_passages"`
	Questions            []*QuestionResult `json:"questions"`
	Anonymous            bool              `json:"anonymous"`
	Participants         []*UserGet        `json:"participants,omitempty"`
}

func (form *FormResult) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
	form.Description = sanitizer.Sanitize(form.Description)
	form.Author.Sanitize(sanitizer)
	for _, question := range form.Questions {
		question.Sanitize(sanitizer)
	}

	for _, user := range form.Participants {
		user.Sanitize(sanitizer)
	}
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
	FormID     int64          `json:"form_id"`
	UserID     sql.NullInt64  `json:"user_id" db:"user_id"`
	Username   string         `json:"username" db:"username"`
	FirstName  string         `json:"first_name" db:"first_name"`
	LastName   string         `json:"last_name" db:"last_name"`
	Email      string         `json:"email" db:"email"`
	Gender     string         `json:"gender" db:"gender"`
	Birthday   sql.NullString `json:"birthday" db:"birthday"`
	QuestionID int64          `json:"question_id" db:"question_id"`
	AnswerText string         `json:"answer_text" db:"answer_text"`
}

type FormPassageGet struct {
	ID          *int64      `json:"id" db:"id"`
	Title       string      `json:"title" validate:"required" db:"title"`
	Description *string     `json:"description" db:"description"`
	UserID      int64       `json:"user_id" db:"user_id"`
	Author      *UserGet    `json:"author"`
	FinishedAt  time.Time   `json:"finished_at" db:"finished_at"`
	Questions   []*Question `json:"questions" validate:"required"`
}

func (form *FormPassageGet) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
	if form.Description != nil {
		*form.Description = sanitizer.Sanitize(*form.Description)
	}
	form.Author.Sanitize(sanitizer)
	for _, question := range form.Questions {
		question.Sanitize(sanitizer)
	}
}

type FormPassageTitle struct {
	ID         int64     `json:"id" validate:"required" db:"id"`
	Title      string    `json:"title" validate:"required" db:"title"`
	FinishedAt time.Time `json:"finished_at" validate:"required" db:"finished_at"`
}

func (form *FormPassageTitle) Sanitize(sanitizer *bluemonday.Policy) {
	form.Title = sanitizer.Sanitize(form.Title)
}

type FormPassageList struct {
	CollectionResponse
	Forms []*FormPassageTitle `json:"forms" validate:"required"`
}

func (forms *FormPassageList) Sanitize(sanitizer *bluemonday.Policy) {
	for _, form := range forms.Forms {
		form.Sanitize(sanitizer)
	}
}
