package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/database"
	"go-form-hub/internal/repository"
	"strings"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepositoryFindByID(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewSessionDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id1 := "this-is-uuid"
		rows := mock.NewRows([]string{"id", "user_id", "created_at"}).
			AddRow(id1, int64(1), time.Now().UTC())
		mock.ExpectQuery(fmt.Sprintf(`^SELECT .* FROM %s.session WHERE id = \$1$`, schema)).
			WithArgs(id1).
			WillReturnRows(rows)

		mock.ExpectCommit()

		session, err := repo.FindByID(context.Background(), id1)
		if err != nil {
			t.Logf("failed to find_by_id form: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id1, session.SessionID)
	})
}

func TestSessionRepositoryFindByUserID(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewSessionDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id := int64(1)
		rows := mock.NewRows([]string{"id", "user_id", "created_at"}).
			AddRow("this-is-uuid", id, time.Now().UTC())
		mock.ExpectQuery(fmt.Sprintf(`^SELECT .* FROM %s.session WHERE user_id = \$1 LIMIT 1$`, schema)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectCommit()

		session, err := repo.FindByUserID(context.Background(), int64(1))
		if err != nil {
			t.Logf("failed to find_by_user_id form: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id, session.UserID)
	})
}

func TestSessionRepositoryInsert(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewSessionDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		session := &repository.Session{
			SessionID: "uuid",
			UserID:    int64(1),
			CreatedAt: time.Now().UTC(),
		}

		mock.ExpectExec(fmt.Sprintf(`^INSERT INTO %s.session (.*) VALUES (.*)$`, schema)).
			WithArgs(session.SessionID, session.UserID, session.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err = repo.Insert(context.Background(), session)
		if err != nil {
			t.Logf("failed to insert session: %e", err)
			t.FailNow()
		}
	})
}

func TestSessionRepositoryDelete(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewSessionDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id := "uuid-string"
		mock.ExpectExec(fmt.Sprintf(`^DELETE FROM %s.session WHERE id = \$1$`, schema)).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err = repo.Delete(context.Background(), id)
		if err != nil {
			t.Logf("failed to delete session: %e", err)
			t.FailNow()
		}
	})
}
