package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Masterminds/squirrel"
	"go-form-hub/internal/database"
)

type URLMapping struct {
	LongURL  string `db:"long_url"`
	ShortURL string `db:"short_url"`
}

type databaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) *databaseRepository {
	return &databaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *databaseRepository) Insert(ctx context.Context, url *URLMapping) (string, error) {
	query, args, err := r.builder.Insert("ShortURL").
		Columns("ShortURL", "LongURL").
		Values(url.ShortURL, url.LongURL).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("shortener_repository save_short_url failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("shortener_repository insert failed to begin transaction: %e", err)
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
		err = fmt.Errorf("shortener_repository insert failed to execute query: %e", err)
		return "", err
	}

	var keyURL string
	err = row.Scan(&keyURL)
	if err != nil {
		return "", fmt.Errorf("shortener_repository insert failed to return key_url: %e", err)
	}

	return keyURL, nil
}

func (r *databaseRepository) RedirectHandler(w http.ResponseWriter, req *http.Request) {
	shortURL := req.URL.Path[len("/redirect/"):]
	if shortURL == "" {
		http.Error(w, "Short URL not provided", http.StatusBadRequest)
		return
	}

	longURL, err := r.GetLongURL(req.Context(), shortURL)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, req, longURL, http.StatusFound)
}

func (r *databaseRepository) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	var longURL string
	query, args, err := r.builder.Select("LongURL").
		From("ShortURL").
		Where(squirrel.Eq{"ShortURL": shortURL}).
		Limit(1).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("shortener_repository get_long_url failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("shortener_repository get_long_url failed to begin transaction: %v", err)
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	err = tx.QueryRow(ctx, query, args...).Scan(&longURL)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("short URL not found")
	} else if err != nil {
		return "", fmt.Errorf("shortener_repository get_long_url failed to execute query: %v", err)
	}

	return longURL, nil
}