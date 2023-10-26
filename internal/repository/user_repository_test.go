package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v3"
)

var (
	builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

func TestRepositoryInsert(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Logf("failed to create mock: %e", err)
		t.FailNow()
	}

	schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	connPool := database.NewConnPool(mock, schema)
	repo := repository.NewUserDatabaseRepository(connPool, builder)

	mock.ExpectBegin()
	mock.ExpectQuery(fmt.Sprintf("^INSERT INTO %s.user (.*) VALUES (.*) RETURNING id", schema)).
		WithArgs("username", "first_name", "last_name", "password", "email").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(
			int64(1),
		))

	mock.ExpectCommit()

	u := repository.User{
		Username:  "username",
		FirstName: "first_name",
		LastName:  "last_name",
		Password:  "password",
		Email:     "email",
	}
	if _, err := repo.Insert(context.Background(), &u); err != nil {
		t.Logf("failed to insert into user repository: %e", err)
		t.FailNow()
	}
}
