package repository

type Answer struct {
	ID         int64  `db:"id"`
	AnswerText string `db:"answer_text"`
	QuestionID int64  `db:"question_id"`
}
