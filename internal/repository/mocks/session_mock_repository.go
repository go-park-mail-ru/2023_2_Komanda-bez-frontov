package repository

import (
	"context"
	"math/rand"
	"go-form-hub/internal/repository"
	"sync"
)

type sessionMockRepository struct {
	mockDB *sync.Map
}

// NewUserMockRepository creates a new instance of the UserMockRepository struct.
//
// It returns a UserRepository interface.
func NewSessionMockRepository() repository.SessionRepository {
	return &sessionMockRepository{
		mockDB: &sync.Map{},
	}
}

// FindAll retrieves all users from the form mock repository.
//
// It takes a context.Context as its parameter and returns a slice of users and an error.
func (r *sessionMockRepository) FindAll(ctx context.Context) ([]*repository.Session, error) {
	sessions := []*repository.Session{}
	r.mockDB.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*repository.Session))
		return true
	})

	return sessions, nil
}

// FindByID retrieves a form from the formMockRepository based on its title.
//
// ctx - The context.Context object for managing the request lifecycle.
// title - The title of the form to search for.
// Returns a pointer to the form object if found, otherwise returns an error.
func (r *sessionMockRepository) FindByID(ctx context.Context, id string) (*repository.Session, error) {
	if session, ok := r.mockDB.Load(id); ok {
		return session.(*repository.Session), nil
	}

	return nil, nil
}

// Delete deletes user from the userMockRepository.
//
// It takes a context.Context and a form as parameters.
// It returns an error.
func (r *sessionMockRepository) Delete(ctx context.Context, id string) error {
	r.mockDB.Delete(id)
	return nil
}

// Insert inserts a form into the formMockRepository.
//
// It takes a context.Context and a form as parameters.
// It returns an error.
func (r *sessionMockRepository) Insert(ctx context.Context, username string) (string, error) {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	
	SID := make([]rune, 32)
	for i := range SID {
		SID[i] = letterRunes[rand.Intn(len(letterRunes))]
	}	

	r.mockDB.Store(string(SID), username)
	return string(SID), nil
}

