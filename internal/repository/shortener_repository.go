package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"

	"go-form-hub/internal/database"

	"github.com/Masterminds/squirrel"
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
	shortURL, err := generateRandomString(8)
	if err != nil {
		return "", err
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

	query, args, err := r.builder.Insert(fmt.Sprintf("%s.url", r.db.GetSchema())).
		Columns("long_url", "short_url").
		Values(url.LongURL, shortURL).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("shortener_repository save_short_url failed to build query: %v", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("shortener_repository insert failed to execute query: %e", err)
	}

	return shortURL, nil
}


func generateRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", fmt.Errorf("failed to generate random string: %v", err)
    }

    randomString := base64.URLEncoding.EncodeToString(bytes)

    return randomString[:length], nil
}


func (r *databaseRepository) RedirectHandler(w http.ResponseWriter, req *http.Request) {
	shortURL := req.URL.Path[len("/123"):]
	if shortURL == "" {
		http.Error(w, "Short URL not provided", http.StatusBadRequest)
		return
	}

	longURL, err := r.GetLongURL(req.Context(), shortURL)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, req, longURL[28:], http.StatusFound)
}

func (r *databaseRepository) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	var longURL string
	query, args, err := r.builder.Select("long_url").
		From(fmt.Sprintf("%s.url", r.db.GetSchema())).
		Where(squirrel.Eq{"short_url": shortURL}).
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
