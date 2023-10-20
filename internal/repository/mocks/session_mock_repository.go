package repository

import (
	"context"
	"fmt"
	"go-form-hub/internal/repository"
	"sync"
)

type sessionMockRepository struct {
	mockDB *sync.Map
}

func NewSessionMockRepository() repository.SessionRepository {
	return &sessionMockRepository{
		mockDB: &sync.Map{},
	}
}

func (r *sessionMockRepository) FindAll(_ context.Context) ([]*repository.Session, error) {
	sessions := []*repository.Session{}
	r.mockDB.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*repository.Session))
		return true
	})

	return sessions, nil
}

func (r *sessionMockRepository) FindByID(_ context.Context, sessionID string) (*repository.Session, error) {
	if session, ok := r.mockDB.Load(sessionID); ok {
		return session.(*repository.Session), nil
	}
	return nil, nil
}

func (r *sessionMockRepository) FindByUsername(_ context.Context, username string) (*repository.Session, error) {
	var session *repository.Session
	r.mockDB.Range(func(key, value interface{}) bool {
		currSession := value.(*repository.Session)
		if currSession.Username == username {
			session = currSession
			return false
		}
		return true
	})

	return session, nil
}

func (r *sessionMockRepository) FindByUserID(_ context.Context, id string) (*repository.Session, error) {
	var session *repository.Session
	r.mockDB.Range(func(key, value interface{}) bool {
		currSession := value.(*repository.Session)
		if currSession.UserID == id {
			session = currSession
			return false
		}
		return true
	})

	return session, nil
}

func (r *sessionMockRepository) Delete(_ context.Context, sessionID string) error {
	r.mockDB.Delete(sessionID)
	return nil
}

func (r *sessionMockRepository) Insert(_ context.Context, session *repository.Session) error {
	if session == nil {
		return fmt.Errorf("session_repository insert session is nil")
	}

	r.mockDB.Store(session.SessionID, session)
	return nil
}
