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

func TestFormRepositoryFindAll(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewFormDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id1 := int64(1)
		id2 := int64(2)
		rows := mock.NewRows([]string{"f.id", "f.title", "f.author_id", "f.created_at", "u.id", "u.username", "u.first_name", "u.last_name", "u.email"}).
			AddRow(&id1, "title1", int64(1), time.Now().UTC(), int64(1), "username1", "first_name1", "last_name1", "email1").
			AddRow(&id2, "title2", int64(1), time.Now().UTC(), int64(1), "username1", "first_name1", "last_name1", "email1")
		mock.ExpectQuery(fmt.Sprintf("^SELECT .* FROM %s.form as f JOIN %s.user as u ON f.author_id = u.id$", schema, schema)).
			WillReturnRows(rows)

		mock.ExpectCommit()

		forms, authors, err := repo.FindAll(context.Background())
		if err != nil {
			t.Logf("failed to find_all forms: %e", err)
			t.FailNow()
		}
		assert.Equal(t, 2, len(forms))
		assert.Equal(t, 1, len(authors))
	})
}

func TestFormRepositoryFindByID(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewFormDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id1 := int64(1)
		rows := mock.NewRows([]string{"f.id", "f.title", "f.author_id", "f.created_at", "u.id", "u.username", "u.first_name", "u.last_name", "u.email"}).
			AddRow(&id1, "title1", int64(1), time.Now().UTC(), int64(1), "username1", "first_name1", "last_name1", "email1")
		mock.ExpectQuery(fmt.Sprintf(`^SELECT .* FROM %s.form as f JOIN %s.user as u ON f.author_id = u.id WHERE f.id = \$1$`, schema, schema)).
			WithArgs(id1).
			WillReturnRows(rows)

		mock.ExpectCommit()

		form, author, err := repo.FindByID(context.Background(), id1)
		if err != nil {
			t.Logf("failed to find_by_id form: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id1, *form.ID)
		assert.Equal(t, "username1", author.Username)
	})
}

func TestFormRepositoryInsert(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))

		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewFormDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		form := &repository.Form{
			Title:     "title1",
			AuthorID:  int64(1),
			CreatedAt: time.Now().UTC(),
		}

		id := int64(1)
		rows := mock.NewRows([]string{"id"}).AddRow(id)
		mock.ExpectQuery(fmt.Sprintf(`^INSERT INTO %s.form (.*) VALUES (.*) RETURNING id$`, schema)).
			WithArgs(form.Title, form.AuthorID, form.CreatedAt).
			WillReturnRows(rows)

		mock.ExpectCommit()

		newID, err := repo.Insert(context.Background(), form)
		if err != nil || newID == nil {
			t.Logf("failed to insert form: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id, *newID)
	})
}

func TestFormRepositoryUpdate(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))

		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewFormDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		id := int64(1)
		form := &repository.Form{
			Title:     "title1",
			AuthorID:  int64(1),
			CreatedAt: time.Now().UTC(),
			ID:        &id,
		}

		rows := mock.NewRows([]string{"id", "title", "created_at"}).AddRow(&id, form.Title, form.CreatedAt)

		mock.ExpectQuery(fmt.Sprintf(`UPDATE %s.form SET title = \$1 WHERE id = \$2 RETURNING id, title, created_at`, schema)).
			WithArgs(form.Title, id).
			WillReturnRows(rows)

		mock.ExpectCommit()

		updatedForm, err := repo.Update(context.Background(), id, form)
		if err != nil {
			t.Logf("failed to update form: %e", err)
			t.FailNow()
		}

		assert.Equal(t, id, *updatedForm.ID)
		assert.Equal(t, form.Title, updatedForm.Title)
	})
}

func TestFormRepositoryDelete(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Logf("failed to create mock: %e", err)
			t.FailNow()
		}

		schema := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))

		connPool := database.NewConnPool(mock, schema)
		repo := repository.NewFormDatabaseRepository(connPool, builder)

		mock.ExpectBegin()

		mock.ExpectExec(fmt.Sprintf(`^DELETE FROM %s.form WHERE id = \$1$`, schema)).
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
