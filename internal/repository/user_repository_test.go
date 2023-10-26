package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/config"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
)

func TestRepository(t *testing.T) {
	schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	dbURL := fmt.Sprintf("postgres://nofronts:nofronts@localhost:5432/nofronts_dev?sslmode=disable&search_path=%s", schema)

	cfg := &config.Config{
		DatabaseURL:                 dbURL,
		DatabaseMaxConnections:      40,
		DatabaseAcquireTimeout:      10,
		DatabaseMigrationsDir:       "../../db/migrations",
		DatabaseConnectMaxRetries:   5,
		DatabaseConnectRetryTimeout: 1 * time.Second,
	}

	db, err := database.ConnectDatabaseWithRetry(cfg)
	if err != nil {
		t.FailNow()
	}

	defer func() {
		_, _ = db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schema))
		db.Close()
	}()

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_, err = database.Migrate(db, cfg, builder)
	if err != nil {
		t.FailNow()
	}

	repo := repository.NewUserDatabaseRepository(db, builder)

	_, err = repo.Insert(context.Background(), &repository.User{
		FirstName: "admin",
		Password:  "admin",
		Email:     "admin",
	})
	if err != nil {
		t.FailNow()
	}
}
