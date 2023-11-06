package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type questionRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewQuestionRepository(db database.ConnPool, builder squirrel.StatementBuilderType) QuestionRepository {
	return &questionRepository{
		db:      db,
		builder: builder,
	}
}

type Question struct {
	ID          int64   `db:"id"`
	FormID      int64   `db:"form_id"`
	Title       *string `db:"title"`
	Description *string `db:"text"`
	Type        string  `db:"type"`
	Shuffle     bool    `db:"shuffle"`
}

func (r *questionRepository) getTableName() string {
	return fmt.Sprintf("%s.question", r.db.GetSchema())
}

func (r *questionRepository) BatchInsert(ctx context.Context, questions []*Question, tx pgx.Tx) ([]int64, error) {
	return nil, nil
}

func (r *questionRepository) Delete(ctx context.Context, ids []int64, tx pgx.Tx) error {
	return nil
}
