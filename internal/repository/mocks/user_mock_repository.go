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

// NewUserMockRepository creates a new instance of the UserMockRepository struct.
//
// It returns a UserRepository interface.
func NewUserMockRepository() repository.UserRepository {
	return &userMockRepository{
		mockDB: &sync.Map{},
	}
}

// FindAll retrieves all users from the form mock repository.
//
// It takes a context.Context as its parameter and returns a slice of users and an error.
func (r *userMockRepository) FindAll(ctx context.Context) ([]*repository.User, error) {
	users := []*repository.User{}
	r.mockDB.Range(func(key, value interface{}) bool {
		users = append(users, value.(*repository.User))
		return true
	})

	return users, nil
}

// FindByName retrieves a form from the formMockRepository based on its title.
//
// ctx - The context.Context object for managing the request lifecycle.
// title - The title of the form to search for.
// Returns a pointer to the form object if found, otherwise returns an error.
func (r *userMockRepository) FindByUsername(ctx context.Context, name string) (*repository.User, error) {
	if user, ok := r.mockDB.Load(name); ok {
		return user.(*repository.User), nil
	}

	return nil, nil
}

// Delete deletes user from the userMockRepository.
//
// It takes a context.Context and a form as parameters.
// It returns an error.
func (r *userMockRepository) Delete(ctx context.Context, name string) error {
	r.mockDB.Delete(name)
	return nil
}

// Insert inserts a form into the formMockRepository.
//
// It takes a context.Context and a form as parameters.
// It returns an error.
func (r *userMockRepository) Insert(ctx context.Context, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository insert user is nil")
	}

	r.mockDB.Store(user.Username, user)
	return nil
}

// Update updates user in the user repository.
//
// It takes a context and an user as parameters.
// It returns an error.
func (r *userMockRepository) Update(ctx context.Context, user *repository.User) error {
	if user == nil {
		return fmt.Errorf("user_repository update user is nil")
	}

	r.mockDB.Store(user.Username, user)
	return nil
}
