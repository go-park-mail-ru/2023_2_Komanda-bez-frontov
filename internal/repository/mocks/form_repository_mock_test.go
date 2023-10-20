package repository_test

import (
	"context"
	"fmt"
	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	repositorymock "go-form-hub/internal/repository/mocks"
	"reflect"
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

	err = repo.Insert(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestFormRepositoryFindByTitleUpdateDelete(t *testing.T) {
	t.Parallel()
	repo := repositorymock.NewFormMockRepository()

	title := "test"
	err := repo.Insert(context.Background(), &repository.Form{
		Title: title,
	})
	if !assert.Nil(t, err) {
		t.Logf("failed to insert into form repository: %e", err)
		t.FailNow()
	}

	form, err := repo.FindByTitle(context.Background(), title)
	if !assert.Nil(t, err) || !assert.NotNil(t, form) {
		t.Logf("failed to find form in form repository: %e", err)
		t.FailNow()
	}

	assert.Equal(t, title, form.Title)

	newTitle := "new-title"
	err = repo.Update(context.Background(), title, &repository.Form{
		Title: newTitle,
	})
	if !assert.Nil(t, err) {
		t.Logf("failed to update form in form repository: %e", err)
		t.FailNow()
	}

	form, err = repo.FindByTitle(context.Background(), title)
	if !assert.Nil(t, err) || !assert.Nil(t, form) {
		t.Logf("failed to find form in form repository: %e", err)
		t.FailNow()
	}

	form, err = repo.FindByTitle(context.Background(), newTitle)
	if !assert.Nil(t, err) || !assert.NotNil(t, form) {
		t.Logf("failed to find form in form repository: %e", err)
		t.FailNow()
	}

	assert.Equal(t, newTitle, form.Title)

	err = repo.Delete(context.Background(), newTitle)
	if !assert.Nil(t, err) {
		t.Logf("failed to delete form in form repository: %e", err)
		t.FailNow()
	}

	form, err = repo.FindByTitle(context.Background(), newTitle)
	if !assert.Nil(t, err) || !assert.Nil(t, form) {
		t.Logf("failed to find form in form repository: %e", err)
		t.FailNow()
	}

	err = repo.Update(context.Background(), "abcd", nil)
	assert.NotNil(t, err)
}

func TestFormRepositoryToModelFromModel(t *testing.T) {
	t.Parallel()
	repo := repositorymock.NewFormMockRepository()

	title := "test"
	formDB := &repository.Form{
		Title: title,
	}

	formModel := repo.ToModel(formDB)

	m := &model.Form{
		Title: title,
	}
	if !reflect.DeepEqual(formModel, m) {
		t.Logf("models are not equal")
		t.FailNow()
	}

	r := repo.FromModel(m)
	if !reflect.DeepEqual(r, formDB) {
		t.Logf("datasbase contents is not equal")
		t.FailNow()
	}
}
