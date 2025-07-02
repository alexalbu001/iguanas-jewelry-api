package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

//storing state in session and session id in cookie

type Session struct {
	UserID    string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Expiry    time.Time `json:"expiry"`
}

type SessionStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

func (s *SessionStore) CreateSession(userID string, email string) string {

	newSession := Session{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		Expiry:    time.Now().Add(24 * time.Hour),
	}

	sessionID := uuid.New().String()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = newSession
	return sessionID
}

func (s *SessionStore) GetSession(sessionID string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, fmt.Errorf("session ID does not exist: %s", sessionID)
	}
	if value.Expiry.Before(time.Now()) { //session expired
		return Session{}, fmt.Errorf("session expired")
	}
	return value, nil
}

func (s *SessionStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	delete(s.sessions, sessionID)
	return nil
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]Session),
		mu:       sync.RWMutex{},
	}
}
