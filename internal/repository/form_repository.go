package repository

import (
	"context"
	"go-form-hub/internal/model"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

type Form struct {
	Title     string    `db:"title"`
	ID        int64     `db:"id"`
	AuthorID  int64     `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
}

type formDatabaseRepository struct {
	conn    *pgx.ConnPool
	builder squirrel.StatementBuilderType
}

func NewFormDatabaseRepository(conn *pgx.ConnPool, builder squirrel.StatementBuilderType) FormRepository {
	return &formDatabaseRepository{
		conn:    conn,
		builder: builder,
	}
}

func (f *formDatabaseRepository) FindAll(ctx context.Context) ([]*Form, error) {
	return nil, nil
}

func (f *formDatabaseRepository) FindByTitle(ctx context.Context, title string) (*Form, error) {
	return nil, nil
}

func (f *formDatabaseRepository) Insert(ctx context.Context, form *Form) error {
	return nil
}

func (f *formDatabaseRepository) Update(ctx context.Context, title string, form *Form) error {
	return nil
}

func (f *formDatabaseRepository) Delete(ctx context.Context, title string) error {
	return nil
}

func (f *formDatabaseRepository) ToModel(form *Form) *model.Form {
	return nil
}

func (f *formDatabaseRepository) FromModel(form *model.Form) *Form {
	return nil
}
