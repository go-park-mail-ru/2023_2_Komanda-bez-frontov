package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	"sync"
)

type userMockRepository struct {
	mockDB *sync.Map
}

func NewUserMockRepository() repository.UserRepository {
	return &userMockRepository{
		mockDB: &sync.Map{},
	}
}

func (r *userMockRepository) FindAll(_ context.Context) ([]*repository.User, error) {
	users := []*repository.User{}
	r.mockDB.Range(func(key, value interface{}) bool {
		users = append(users, value.(*repository.User))
		return true
	})

	return users, nil
}

func (r *userMockRepository) FindByID(_ context.Context, id string) (*repository.User, error) {
	if user, ok := r.mockDB.Load(id); ok {
		return user.(*repository.User), nil
	}

	return nil, nil
}

func (r *userMockRepository) FindByUsername(_ context.Context, username string) (*repository.User, error) {
	var user *repository.User
	r.mockDB.Range(func(key, value interface{}) bool {
		if value.(*repository.User).Username == username {
			user = value.(*repository.User)
			return false
		}

		return true
	})

	return user, nil
}

func (r *userMockRepository) Delete(_ context.Context, id string) error {
	r.mockDB.Delete(id)
	return nil
}

func (r *userMockRepository) Insert(_ context.Context, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository insert user is nil")
	}

	r.mockDB.Store(user.ID, user)
	return nil
}

func (r *userMockRepository) Update(_ context.Context, id string, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository update user is nil")
	}

	user.ID = id
	r.mockDB.Store(id, user)
	return nil
}
