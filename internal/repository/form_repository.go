package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"
	"go-form-hub/internal/model"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	// "github.com/rs/zerolog/log"
)

type Form struct {
	Title     string    `db:"title"`
	ID        int64     `db:"id"`
	AuthorID  int64     `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
}

var (
	selectFields = []string{
		"f.id",
		"f.title",
		"f.created_at",
		"f.author_id",
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"q.id",
		"q.title",
		"q.text",
		"q.type",
		"q.shuffle",
		"a.id",
		"a.answer_text",
	}
)

type formDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewFormDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) FormRepository {
	return &formDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *formDatabaseRepository) FindAll(ctx context.Context) (forms []*model.Form, err error) {
	query, _, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to execute query: %e", err)
	}

	return r.fromRows(rows)
}

func (r *formDatabaseRepository) FindAllByUser(ctx context.Context, username string) (forms []*model.Form, err error) {
	query, args, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"u.username": username}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to execute query: %e", err)
	}

	return r.fromRows(rows)
}

func (r *formDatabaseRepository) FindByID(ctx context.Context, id int64) (form *model.Form, err error) {
	query, args, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"f.id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to build query: %e", err)
	}

	// log.Info().Msgf("query: %s", query)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to execute query: %e", err)
	}

	forms, err := r.fromRows(rows)
	if len(forms) == 0 {
		return nil, nil
	}

	return forms[0], err
}

func (r *formDatabaseRepository) Insert(ctx context.Context, form *model.Form, tx pgx.Tx) (*model.Form, error) {
	var err error

	if tx == nil {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
		}
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	formQuery, args, err := r.builder.
		Insert(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Columns("title", "author_id", "created_at").
		Values(form.Title, form.Author.ID, form.CreatedAt).
		Suffix("RETURNING id").
		ToSql()
	err = tx.QueryRow(ctx, formQuery, args...).Scan(&form.ID)
	if err != nil {
		return nil, err
	}

	questionBatch := &pgx.Batch{}
	questionQuery := r.builder.
		Insert(fmt.Sprintf("%s.question", r.db.GetSchema())).
		Columns("title", "text", "type", "shuffle", "form_id").
		Suffix("RETURNING id")

	for _, question := range form.Questions {
		q, args, err := questionQuery.Values(question.Title, question.Description, question.Type, question.Shuffle, form.ID).ToSql()
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

	for _, question := range form.Questions {
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
	for _, question := range form.Questions {
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

	return form, nil
}

func (r *formDatabaseRepository) Update(ctx context.Context, id int64, form *model.Form) (result *model.Form, err error) {
	query, args, err := r.builder.Update(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Set("title", form.Title).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, title, created_at").ToSql()
	if err != nil {
		return nil, fmt.Errorf("form_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository update failed to begin transaction: %e", err)
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
		return nil, fmt.Errorf("form_repository update failed to execute query: %e", err)
	}

	return form, nil
}

func (r *formDatabaseRepository) Delete(ctx context.Context, id int64) (err error) {
	query, args, err := r.builder.Delete(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("form_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("form_repository delete failed to begin transaction: %e", err)
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
		return fmt.Errorf("form_repository delete failed to execute query: %e", err)
	}

	return nil
}

func (r *formDatabaseRepository) fromRows(rows pgx.Rows) ([]*model.Form, error) {
	defer func() {
		rows.Close()
	}()

	formMap := map[int64]*model.Form{}
	questionsByFormID := map[int64][]*model.Question{}
	answersByQuestionID := map[int64][]*model.Answer{}

	questionWasAppended := map[int64]bool{}

	for rows.Next() {
		info, err := r.fromRow(rows)
		if err != nil {
			return nil, err
		}

		if info.form == nil {
			continue
		}

		if _, ok := formMap[info.form.ID]; !ok {
			formMap[info.form.ID] = &model.Form{
				ID:        &info.form.ID,
				Title:     info.form.Title,
				CreatedAt: info.form.CreatedAt,
				Author: &model.UserGet{
					ID:        info.author.ID,
					Username:  info.author.Username,
					FirstName: info.author.FirstName,
					LastName:  info.author.LastName,
					Email:     info.author.Email,
					Avatar:    info.author.Avatar,
				},
			}
		}

		if _, ok := questionWasAppended[info.question.ID]; !ok {
			questionsByFormID[info.form.ID] = append(questionsByFormID[info.form.ID], &model.Question{
				ID:          &info.question.ID,
				Title:       info.question.Title,
				Description: info.question.Text,
				Type:        info.question.Type,
				Shuffle:     info.question.Shuffle,
			})
			questionWasAppended[info.question.ID] = true
		}

		if _, ok := answersByQuestionID[info.question.ID]; !ok {
			answersByQuestionID[info.question.ID] = make([]*model.Answer, 0, 1)
		}

		answersByQuestionID[info.question.ID] = append(answersByQuestionID[info.question.ID], &model.Answer{
			ID:   &info.answer.ID,
			Text: info.answer.AnswerText,
		})
	}

	forms := make([]*model.Form, 0, len(formMap))

	for _, form := range formMap {
		form.Questions = questionsByFormID[*form.ID]
		for _, question := range form.Questions {
			question.Answers = answersByQuestionID[*question.ID]
		}
		forms = append(forms, form)
	}

	return forms, nil
}

type fromRowReturn struct {
	form     *Form
	author   *User
	question *Question
	answer   *Answer
}

func (r *formDatabaseRepository) fromRow(row pgx.Row) (*fromRowReturn, error) {
	form := &Form{}
	author := &User{}
	question := &Question{}
	answer := &Answer{}

	err := row.Scan(
		&form.ID,
		&form.Title,
		&form.CreatedAt,
		&form.AuthorID,
		&author.ID,
		&author.Username,
		&author.FirstName,
		&author.LastName,
		&author.Email,
		&question.ID,
		&question.Title,
		&question.Text,
		&question.Type,
		&question.Shuffle,
		&answer.ID,
		&answer.AnswerText,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("form_repository failed to scan row: %e", err)
	}

	return &fromRowReturn{form, author, question, answer}, nil
}
