package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"
	"go-form-hub/internal/model"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Form struct {
	Title     string    `db:"title"`
	ID        *int64    `db:"id"`
	AuthorID  int64     `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
}

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

func (r *formDatabaseRepository) FindAll(ctx context.Context) (forms []*Form, err error) {
	query, _, err := r.builder.Select("id", "title", "author_id", "created_at").
		From(fmt.Sprintf("%s.form", r.db.GetSchema())).ToSql()
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

	forms, err = r.fromRows(rows)

	return forms, err
}

func (r *formDatabaseRepository) FindByID(ctx context.Context, id int64) (form *Form, err error) {
	query, args, err := r.builder.Select("id", "title", "author_id", "created_at").
		From(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Where(squirrel.Eq{"id": id}).ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to build query: %e", err)
	}

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

	row := tx.QueryRow(ctx, query, args...)
	if row == nil {
		err = fmt.Errorf("form_repository find_by_title failed to execute query: %e", err)
	}

	form, err = r.fromRow(row)

	return form, err
}

func (r *formDatabaseRepository) Insert(ctx context.Context, form *Form) (*int64, error) {
	query, args, err := r.builder.Insert(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Columns("title", "author_id", "created_at").
		Values(form.Title, form.AuthorID, time.Now()).
		Suffix("RETURNING id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("form_repository insert failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, query, args...)
	if row == nil {
		return nil, fmt.Errorf("form_repository insert failed to execute query: %e", err)
	}

	var id int64
	err = row.Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("form_repository insert failed to return id: %e", err)
	}

	return &id, nil
}

func (r *formDatabaseRepository) Update(ctx context.Context, id int64, form *Form) (err error) {
	query, args, err := r.builder.Update(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Set("title", form.Title).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("form_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("form_repository update failed to begin transaction: %e", err)
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
		return fmt.Errorf("form_repository update failed to execute query: %e", err)
	}

	return nil
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

func (r *formDatabaseRepository) ToModel(form *Form) *model.Form {
	return &model.Form{
		ID:        form.ID,
		Title:     form.Title,
		AuthorID:  form.AuthorID,
		CreatedAt: form.CreatedAt,
	}
}

func (r *formDatabaseRepository) FromModel(form *model.Form) *Form {
	return &Form{
		ID:        form.ID,
		Title:     form.Title,
		AuthorID:  form.AuthorID,
		CreatedAt: form.CreatedAt,
	}
}

func (r *formDatabaseRepository) fromRows(rows pgx.Rows) ([]*Form, error) {
	defer func() {
		rows.Close()
	}()

	forms := []*Form{}

	for rows.Next() {
		form := &Form{}

		err := rows.Scan(&form.ID, &form.Title, &form.AuthorID, &form.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
		}

		forms = append(forms, form)
	}

	return forms, nil
}

func (r *formDatabaseRepository) fromRow(row pgx.Row) (*Form, error) {
	form := &Form{}
	err := row.Scan(&form.ID, &form.Title, &form.AuthorID, &form.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("user_repository failed to scan row: %e", err)
	}

	return form, nil
}
