package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	repositorymock "go-form-hub/internal/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRepositoryInsertFindAllDelete(t *testing.T) {
	t.Parallel()

	repo := repositorymock.NewUserMockRepository()

	for i := 0; i < 10; i++ {
		err := repo.Insert(context.Background(), &repository.User{
			Username: fmt.Sprintf("test-username-%d", i),
			ID:       fmt.Sprintf("test-id-%d", i),
			Name:     fmt.Sprintf("test-name-%d", i),
			Email:    fmt.Sprintf("test-email-%d", i),
			Surname:  fmt.Sprintf("test-surname-%d", i),
			Password: fmt.Sprintf("test-password-%d", i),
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into user repository: %e", err)
			t.FailNow()
		}
	}

	err := repo.Insert(context.Background(), nil)
	assert.NotNil(t, err)

	users, err := repo.FindAll(context.Background())
	if !assert.Nil(t, err) {
		t.Logf("failed to find all users in user repository: %e", err)
		t.FailNow()
	}

	assert.Len(t, users, 10)

	for i := 0; i < 10; i++ {
		err := repo.Delete(context.Background(), fmt.Sprintf("test-id-%d", i))
		if !assert.Nil(t, err) {
			t.Logf("failed to delete from user repository: %e", err)
			t.FailNow()
		}
	}

	users, err = repo.FindAll(context.Background())
	if !assert.Nil(t, err) {
		t.Logf("failed to find all users in user repository: %e", err)
		t.FailNow()
	}

	assert.Len(t, users, 0)
}

func TestUserRepositoryFindByFunctionsAndUpdate(t *testing.T) {
	t.Parallel()

	repo := repositorymock.NewUserMockRepository()

	for i := 0; i < 2; i++ {
		err := repo.Insert(context.Background(), &repository.User{
			Username: fmt.Sprintf("test-username-%d", i),
			ID:       fmt.Sprintf("test-id-%d", i),
			Name:     fmt.Sprintf("test-name-%d", i),
			Email:    fmt.Sprintf("test-email-%d", i),
			Surname:  fmt.Sprintf("test-surname-%d", i),
			Password: fmt.Sprintf("test-password-%d", i),
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into user repository: %e", err)
			t.FailNow()
		}
	}

	user, err := repo.FindByID(context.Background(), "test-id-0")
	if !assert.Nil(t, err) || !assert.NotNil(t, user) {
		t.Logf("failed to find user in user repository: %e", err)
	}

	assert.Equal(t, "test-username-0", user.Username)

	user, err = repo.FindByUsername(context.Background(), "test-username-1")
	if !assert.Nil(t, err) || !assert.NotNil(t, user) {
		t.Logf("failed to find user in user repository: %e", err)
	}

	assert.Equal(t, "test-id-1", user.ID)

	err = repo.Update(context.Background(), "test-id-0", &repository.User{
		Username: fmt.Sprintf("test-username-%d", 2),
		Name:     fmt.Sprintf("test-name-%d", 2),
		Email:    fmt.Sprintf("test-email-%d", 2),
		Surname:  fmt.Sprintf("test-surname-%d", 2),
		Password: fmt.Sprintf("test-password-%d", 2),
	})
	if !assert.Nil(t, err) {
		t.Logf("failed to update user in user repository: %e", err)
	}

	user, err = repo.FindByID(context.Background(), "test-id-0")
	if !assert.Nil(t, err) || !assert.NotNil(t, user) {
		t.Logf("failed to find user in user repository: %e", err)
	}

	assert.Equal(t, "test-username-2", user.Username)

	user, err = repo.FindByID(context.Background(), "test-id-123213")
	if !assert.Nil(t, err) || !assert.Nil(t, user) {
		t.FailNow()
	}

	err = repo.Update(context.Background(), "test-id-1", nil)
	assert.NotNil(t, err)
}
