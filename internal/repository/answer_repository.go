package repository

import (
	"context"
	"fmt"

	"go-form-hub/internal/database"
	"go-form-hub/internal/model"

	"github.com/Masterminds/squirrel"
)

type Answer struct {
	ID         int64  `db:"id"`
	AnswerText string `db:"answer_text"`
	QuestionID int64  `db:"question_id"`
}

type answerDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewAnswerDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) AnswerRepository {
	return &answerDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *answerDatabaseRepository) DeleteAllByID(ctx context.Context, ids []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("answer_repository delete failed to begin transaction: %e", err)
	}

	query, args, err := r.builder.
		Delete(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("answer_repository delete failed to build query: %e", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *answerDatabaseRepository) Update(ctx context.Context, id int64, answer *model.Answer) error {
	fmt.Println(id, answer)
	query, args, err := r.builder.Update(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Set("answer_text", answer.Text).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("answer_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("answer_repository update failed to begin transaction: %e", err)
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
		return fmt.Errorf("answer_repository update failed to execute query: %e", err)
	}

	return nil
}

func (r *answerDatabaseRepository) Insert(ctx context.Context, questionID int64, answer *model.Answer) error {
	query, args, err := r.builder.Insert(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Columns("answer_text", "question_id").
		Values(answer.Text, questionID).
		Suffix("RETURNING id").ToSql()
	if err != nil {
		return fmt.Errorf("answer_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("answer_repository update failed to begin transaction: %e", err)
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
		return fmt.Errorf("answer_repository update failed to execute query: %e", err)
	}

	return nil
}

func (r *answerDatabaseRepository) DeleteByQuestionID(ctx context.Context, questionID int64) error {
	query, args, err := r.builder.
		Delete(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Where(squirrel.Eq{"question_id": questionID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("answer_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("answer_repository delete failed to begin transaction: %e", err)
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
