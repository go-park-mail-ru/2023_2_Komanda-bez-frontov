package repository

import "time"

type Answer struct {
	ID         int64  `db:"id"`
	AnswerText string `db:"answer_text"`
	QuestionID int64  `db:"question_id"`
}

type PassageAnswer struct {
	ID         int64     `db:"id"`
	AnswerText string    `db:"answer_text"`
	QuestionID int64     `db:"question_id"`
	UserID     *int64    `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
}
