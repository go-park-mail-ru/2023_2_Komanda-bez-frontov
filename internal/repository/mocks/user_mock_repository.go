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

func (r *userMockRepository) FindAll(ctx context.Context) ([]*repository.User, error) {
	users := []*repository.User{}
	r.mockDB.Range(func(key, value interface{}) bool {
		users = append(users, value.(*repository.User))
		return true
	})

	return users, nil
}

func (r *userMockRepository) FindByUsername(ctx context.Context, name string) (*repository.User, error) {
	if user, ok := r.mockDB.Load(name); ok {
		return user.(*repository.User), nil
	}

	return nil, nil
}

func (r *userMockRepository) Delete(ctx context.Context, name string) error {
	r.mockDB.Delete(name)
	return nil
}

func (r *userMockRepository) Insert(ctx context.Context, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository insert user is nil")
	}

	r.mockDB.Store(user.Username, user)
	return nil
}

func (r *userMockRepository) Update(ctx context.Context, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository update user is nil")
	}

	r.mockDB.Store(user.Username, user)
	return nil
}
