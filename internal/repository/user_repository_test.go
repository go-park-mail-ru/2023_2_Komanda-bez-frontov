package repository_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

var (
	builder     = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	emptyString = (*string)(nil)
)

func TestUserRepositoryFindAll(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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

		rows := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email", "avatar"}).
			AddRow(int64(1), "username1", "first_name1", "last_name1", "password1", "email1", emptyString).
			AddRow(int64(2), "username2", "first_name2", "last_name2", "password2", "email2", emptyString)

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
		rows := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email", "avatar"}).
			AddRow(int64(1), username, "first_name1", "last_name1", "password1", "email1", emptyString)

		mock.ExpectQuery(fmt.Sprintf(`^SELECT (.*) FROM %s.user WHERE username = .* LIMIT 1`, schema)).
			WithArgs(username).
			WillReturnRows(rows)

		mock.ExpectCommit()

		user, err := repo.FindByUsername(context.Background(), username)
		if err != nil {
			t.Logf("failed to find_by_username: %e", err)
			t.FailNow()
		}

		assert.Equal(t, username, user.Username)
	})
}

func TestUserRepositoryFindByEmail(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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

		email := "unique-email"
		rows := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email", "avatar"}).
			AddRow(int64(1), "username", "first_name1", "last_name1", "password1", email, emptyString)

		mock.ExpectQuery(fmt.Sprintf(`^SELECT (.*) FROM %s.user WHERE email = \$1 LIMIT 1`, schema)).
			WithArgs(email).
			WillReturnRows(rows)

		mock.ExpectCommit()

		user, err := repo.FindByEmail(context.Background(), email)
		if err != nil {
			t.Logf("failed to find_by_username: %e", err)
			t.FailNow()
		}

		assert.Equal(t, email, user.Email)
	})
}

func TestUserRepositoryFindByUD(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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

		id := int64(123)
		rows := mock.NewRows([]string{"id", "username", "first_name", "last_name", "password", "email", "avatar"}).
			AddRow(id, "username", "first_name1", "last_name1", "password1", "email", emptyString)

		mock.ExpectQuery(fmt.Sprintf(`^SELECT (.*) FROM %s.user WHERE id = \$1 LIMIT 1`, schema)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectCommit()

		user, err := repo.FindByID(context.Background(), id)
		if err != nil {
			t.Logf("failed to find_by_username: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id, user.ID)
	})
}

func TestUserRepositoryInsert(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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
			WithArgs("username", "first_name", "last_name", "password", "email", emptyString).
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
			Avatar:    emptyString,
		}

		id, err := repo.Insert(context.Background(), &u)
		if err != nil {
			t.Logf("failed to insert into user repository: %e", err)
			t.FailNow()
		}

		assert.Equal(t, int64(1), id)
	})

	t.Run("ErrorBeginTransaction", func(t *testing.T) {
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
			Avatar:    emptyString,
		}

		_, err = repo.Insert(context.Background(), &u)
		if err == nil {
			t.Log("failed to insert into user repository. Error expected")
			t.FailNow()
		}

		expectedErr := fmt.Errorf("user_repository insert failed to begin transaction: %e", beginErr)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("ErrorReturnID", func(t *testing.T) {
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
			WithArgs("username", "first_name", "last_name", "password", "email", emptyString).
			WillReturnError(queryErr)
		mock.ExpectRollback()

		u := repository.User{
			Username:  "username",
			FirstName: "first_name",
			LastName:  "last_name",
			Password:  "password",
			Email:     "email",
			Avatar:    emptyString,
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

func TestUserRepositoryUpdate(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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

		query := fmt.Sprintf(`^UPDATE %s.user SET .* WHERE id = \$7$`, schema)
		mock.ExpectExec(query).
			WithArgs("username", "first_name", "last_name", "password", "email", emptyString, int64(1)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()

		u := repository.User{
			ID:        1,
			Username:  "username",
			FirstName: "first_name",
			LastName:  "last_name",
			Password:  "password",
			Email:     "email",
			Avatar:    nil,
		}

		err = repo.Update(context.Background(), 1, &u)
		if err != nil {
			t.Logf("failed to update user: %e", err)
			t.FailNow()
		}
	})
}

func TestUserRepositoryDelete(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
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

		mock.ExpectExec(fmt.Sprintf(`^DELETE FROM %s.user WHERE id = \$1$`, schema)).
			WithArgs(int64(1)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		mock.ExpectCommit()

		err = repo.Delete(context.Background(), 1)
		if err != nil {
			t.Logf("failed to delete user: %e", err)
			t.FailNow()
		}
	})
}
