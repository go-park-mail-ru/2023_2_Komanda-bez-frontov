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
	"github.com/stretchr/testify/assert"
)

var (
	builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

func TestUserRepositoryFindAll(t *testing.T) {
	t.Run("FindAllUserNoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewUserDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		rows := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email"}).
			AddRow(int64(1), "username1", "first_name1", "last_name1", "password1", "email1").
			AddRow(int64(2), "username2", "first_name2", "last_name2", "password2", "email2")

		mock.ExpectQuery(fmt.Sprintf(`^SELECT (.*) FROM %s.user`, schema)).
			WillReturnRows(rows)

		mock.ExpectCommit()

		users, err := repo.FindAll(context.Background())
		if err != nil {
			t.Logf("failed to find_all users: %e", err)
			t.FailNow()
		}

		assert.Equal(t, 2, len(users))
	})
}

func TestUserRepositoryFindByUsername(t *testing.T) {
	t.Run("FindUserByUsernameNoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewUserDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		username := "unique-username"
		row := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email"}).
			AddRow(int64(1), "username", "first_name1", "last_name1", "password1", "email1")

		mock.ExpectQuery(fmt.Sprintf(`^SELECT (.*) FROM %s.user WHERE username = \$1 LIMIT 1`, schema)).
			WillReturnRows(row)

		mock.ExpectCommit()

		user, err := repo.FindByUsername(context.Background(), username)
		if err != nil {
			t.Logf("failed to find_by_username: %e", err)
			t.FailNow()
		}

		assert.Equal(t, username, user.Username)
	})
}

func TestRepositoryInsert(t *testing.T) {
	t.Run("InserUserNoErrors", func(t *testing.T) {
		t.Parallel()
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

		id, err := repo.Insert(context.Background(), &u)
		if err != nil {
			t.Logf("failed to insert into user repository: %e", err)
			t.FailNow()
		}

		assert.Equal(t, int64(1), id)
	})

	t.Run("InserUserErrorBeginTransaction", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewUserDatabaseRepository(connPool, builder)
		beginErr := fmt.Errorf("some error")

		mock.ExpectBegin().WillReturnError(beginErr)

		u := repository.User{
			Username:  "username",
			FirstName: "first_name",
			LastName:  "last_name",
			Password:  "password",
			Email:     "email",
		}

		_, err = repo.Insert(context.Background(), &u)
		if err == nil {
			t.Log("failed to insert into user repository. Error expected")
			t.FailNow()
		}

		expectedErr := fmt.Errorf("user_repository insert failed to begin transaction: %e", beginErr)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("InsertUserNilRowsReturned", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewUserDatabaseRepository(connPool, builder)

		queryErr := fmt.Errorf("query error")
		mock.ExpectBegin()
		mock.ExpectQuery(fmt.Sprintf("^INSERT INTO %s.user (.*) VALUES (.*) RETURNING id", schema)).
			WithArgs("username", "first_name", "last_name", "password", "email").
			WillReturnError(queryErr)
		mock.ExpectRollback()

		u := repository.User{
			Username:  "username",
			FirstName: "first_name",
			LastName:  "last_name",
			Password:  "password",
			Email:     "email",
		}

		_, err = repo.Insert(context.Background(), &u)
		if err == nil {
			t.Log("failed to insert into user repository. Error expected")
			t.FailNow()
		}

		expectedErr := fmt.Errorf("user_repository insert failed to return id: %e", queryErr)
		assert.Equal(t, expectedErr, err)
	})
}
