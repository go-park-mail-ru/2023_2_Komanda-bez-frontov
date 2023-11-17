package repository

import (
	"context"
	"fmt"

	"go-form-hub/internal/database"
	"go-form-hub/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Question struct {
	ID       int64   `db:"id"`
	FormID   int64   `db:"form_id"`
	Type     int     `db:"type"`
	Title    string  `db:"title"`
	Text     *string `db:"text"`
	Required bool    `db:"required"`
}

type questionDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewQuestionDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) QuestionRepository {
	return &questionDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *questionDatabaseRepository) BatchInsert(ctx context.Context, questions []*model.Question, formID int64) ([]*model.Question, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("session_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	questionBatch := &pgx.Batch{}
	questionQuery := r.builder.
		Insert(fmt.Sprintf("%s.question", r.db.GetSchema())).
		Columns("title", "text", "type", "required", "form_id").
		Suffix("RETURNING id")

	for _, question := range questions {
		q, args, err := questionQuery.Values(question.Title, question.Description, question.Type, question.Required, formID).ToSql()
		if err != nil {
			return nil, err
		}

		questionBatch.Queue(q, args...)
	}
	questionResults := tx.SendBatch(ctx, questionBatch)

	answerBatch := &pgx.Batch{}
	answerQuery := r.builder.
		Insert(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Columns("answer_text", "question_id").
		Suffix("RETURNING id")

	for _, question := range questions {
		questionID := int64(0)
		err = questionResults.QueryRow().Scan(&questionID)
		if err != nil {
			return nil, err
		}
		question.ID = &questionID
		for _, answer := range question.Answers {
			q, args, err := answerQuery.Values(answer.Text, question.ID).ToSql()
			if err != nil {
				return nil, err
			}

			answerBatch.Queue(q, args...)
		}
	}
	questionResults.Close()

	answerResults := tx.SendBatch(ctx, answerBatch)
	for _, question := range questions {
		for _, answer := range question.Answers {
			answerID := int64(0)
			err = answerResults.QueryRow().Scan(&answerID)
			if err != nil {
				return nil, err
			}
			answer.ID = &answerID
		}
	}
	answerResults.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *questionDatabaseRepository) DeleteByFormID(ctx context.Context, formID int64) error {
	query, args, err := r.builder.
		Delete(fmt.Sprintf("%s.question", r.db.GetSchema())).
		Where(squirrel.Eq{"form_id": formID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("session_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("session_repository delete failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)

	return err
}
