package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	repositorymock "go-form-hub/internal/repository/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionRepositoryInsertFindAll(t *testing.T) {
	t.Parallel()

	repo := repositorymock.NewSessionMockRepository()

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("test-%d", i)
		err := repo.Insert(context.Background(), &repository.Session{
			SessionID: s,
			Username:  s,
			UserID:    s,
			CreatedAt: time.Now().UnixMilli(),
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into session repository: %e", err)
			t.FailNow()
		}
	}

	sessions, err := repo.FindAll(context.Background())
	if !assert.Nil(t, err) {
		t.Logf("failed to find all sessions in session repository: %e", err)
		t.FailNow()
	}

	assert.Len(t, sessions, 10)

	err = repo.Insert(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestSessionRepositoryFindByFunctionsAndDelete(t *testing.T) {
	t.Parallel()

	repo := repositorymock.NewSessionMockRepository()

	for i := 0; i < 3; i++ {
		err := repo.Insert(context.Background(), &repository.Session{
			SessionID: fmt.Sprintf("test-id-%d", i),
			Username:  fmt.Sprintf("test-username-%d", i),
			UserID:    fmt.Sprintf("test-user-id-%d", i),
			CreatedAt: time.Now().UnixMilli(),
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into session repository: %e", err)
			t.FailNow()
		}
	}

	session, err := repo.FindByID(context.Background(), "test-id-0")
	if !assert.Nil(t, err) || !assert.NotNil(t, session) {
		t.Logf("failed to find session in session repository: %e", err)
		t.FailNow()
	}

	assert.Equal(t, "test-id-0", session.SessionID)
	assert.Equal(t, "test-username-0", session.Username)
	assert.Equal(t, "test-user-id-0", session.UserID)

	session, err = repo.FindByUserID(context.Background(), "test-user-id-1")
	if !assert.Nil(t, err) || !assert.NotNil(t, session) {
		t.Logf("failed to find session in session repository: %e", err)
		t.FailNow()
	}
	assert.Equal(t, "test-id-1", session.SessionID)
	assert.Equal(t, "test-username-1", session.Username)
	assert.Equal(t, "test-user-id-1", session.UserID)

	session, err = repo.FindByUsername(context.Background(), "test-username-2")
	if !assert.Nil(t, err) || !assert.NotNil(t, session) {
		t.Logf("failed to find session in session repository: %e", err)
		t.FailNow()
	}

	assert.Equal(t, "test-id-2", session.SessionID)
	assert.Equal(t, "test-username-2", session.Username)
	assert.Equal(t, "test-user-id-2", session.UserID)

	for i := 0; i < 3; i++ {
		s := fmt.Sprintf("test-id-%d", i)
		err := repo.Delete(context.Background(), s)
		if !assert.Nil(t, err) {
			t.Logf("failed to delete from session repository: %e", err)
			t.FailNow()
		}
	}

	sessions, err := repo.FindAll(context.Background())
	if !assert.Nil(t, err) {
		t.Logf("failed to find all sessions in session repository: %e", err)
		t.FailNow()
	}

	assert.Len(t, sessions, 0)

	session, err = repo.FindByID(context.Background(), "not-exists")
	if !assert.Nil(t, err) || !assert.Nil(t, session) {
		t.FailNow()
	}

	session, err = repo.FindByUsername(context.Background(), "not-exists")
	if !assert.Nil(t, err) || !assert.Nil(t, session) {
		t.FailNow()
	}
}
