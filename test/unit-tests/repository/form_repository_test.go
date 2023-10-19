package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	repositorymock "go-form-hub/internal/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormRepositoryInsertFindAll(t *testing.T) {
	t.Parallel()

	repo := repositorymock.NewFormMockRepository()

	for i := 0; i < 10; i++ {
		err := repo.Insert(context.Background(), &repository.Form{
			Title: fmt.Sprintf("test-%d", i),
		})
		if !assert.Nil(t, err) {
			t.Logf("failed to insert into form repository: %e", err)
			t.FailNow()
		}
	}

	forms, err := repo.FindAll(context.Background())
	if !assert.Nil(t, err) {
		t.Logf("failed to find all forms in form repository: %e", err)
		t.FailNow()
	}

	assert.Len(t, forms, 10)
}
