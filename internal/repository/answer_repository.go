package repository

type Answer struct {
	ID         int64  `db:"id"`
	QuestionID int64  `db:"question_id"`
	Text       string `db:"text"`
}
