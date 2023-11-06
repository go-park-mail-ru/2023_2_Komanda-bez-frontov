package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type answerRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewAnswerRepository(db database.ConnPool, builder squirrel.StatementBuilderType) QuestionRepository {
	return &answerRepository{
		db:      db,
		builder: builder,
	}
}

type Answer struct {
	ID         int64  `db:"id"`
	QuestionID int64  `db:"answer_id"`
	Text       string `db:"text"`
}

func (r *answerRepository) getTableName() string {
	return fmt.Sprintf("%s.answer", r.db.GetSchema())
}

func (r *answerRepository) BatchInsert(ctx context.Context, answers []*Question, tx pgx.Tx) ([]int64, error) {
	return nil, nil
}

func (r *answerRepository) Delete(ctx context.Context, ids []int64, tx pgx.Tx) error {
	return nil
}
