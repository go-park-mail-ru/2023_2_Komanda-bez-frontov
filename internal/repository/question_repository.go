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

func (r *questionDatabaseRepository) BatchInsert(ctx context.Context, question *model.Question, formID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("session_repository insert failed to begin transaction: %e", err)
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

	q, args, err := questionQuery.Values(question.Title, question.Description, question.Type, question.Required, formID).ToSql()
	if err != nil {
		return err
	}

	questionBatch.Queue(q, args...)

	questionResults := tx.SendBatch(ctx, questionBatch)

	answerBatch := &pgx.Batch{}
	answerQuery := r.builder.
		Insert(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Columns("answer_text", "question_id").
		Suffix("RETURNING id")

	questionID := int64(0)
	err = questionResults.QueryRow().Scan(&questionID)
	if err != nil {
		return err
	}
	question.ID = &questionID
	for _, answer := range question.Answers {
		q, args, err := answerQuery.Values(answer.Text, question.ID).ToSql()
		if err != nil {
			return err
		}

		answerBatch.Queue(q, args...)
	}

	questionResults.Close()

	answerResults := tx.SendBatch(ctx, answerBatch)

	for _, answer := range question.Answers {
		answerID := int64(0)
		err = answerResults.QueryRow().Scan(&answerID)
		if err != nil {
			return err
		}
		answer.ID = &answerID
	}

	answerResults.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
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

func (r *questionDatabaseRepository) DeleteAllByID(ctx context.Context, ids []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("question_repository delete failed to begin transaction: %e", err)
	}

	for _, id := range ids {
		query, args, err := r.builder.
			Delete(fmt.Sprintf("%s.question", r.db.GetSchema())).
			Where(squirrel.Eq{"id": id}).
			ToSql()
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("question_repository delete failed to build query: %e", err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *questionDatabaseRepository) Update(ctx context.Context, id int64, question *model.Question) error {
	query, args, err := r.builder.Update(fmt.Sprintf("%s.question", r.db.GetSchema())).
		Set("title", question.Title).
		Set("text", question.Description).
		Set("type", question.Type).
		Set("required", question.Required).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("question_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("qestion_repository update failed to begin transaction: %e", err)
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
	if err != nil {
		return fmt.Errorf("question_repository update failed to execute query: %e", err)
	}

	return nil
}
