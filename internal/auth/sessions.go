package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//storing state in session and session id in cookie

type Session struct {
	UserID    string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Expiry    time.Time `json:"expiry"`
}

type SessionStore struct {
	// sessions map[string]Session
	// mu       sync.RWMutex
	client *redis.Client
}

func (s *SessionStore) CreateSession(userID string, email string) (string, error) {

	newSession := Session{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		Expiry:    time.Now().Add(24 * time.Hour),
	}

	sessionID := uuid.New().String()
	sessionData, err := json.Marshal(newSession) // Marshall struct to json then pass it to redis
	if err != nil {
		return sessionID, fmt.Errorf("error marshalling struct into json: %w", err)
	}
	err = s.client.Set(context.Background(), sessionID, sessionData, 24*time.Hour).Err()
	if err != nil {
		return sessionID, fmt.Errorf("error creating session: %w", err)
	}
	// s.mu.Lock()
	// defer s.mu.Unlock()
	// s.sessions[sessionID] = newSession
	return sessionID, nil
}

func (s *SessionStore) GetSession(sessionID string) (Session, error) {
	// s.mu.RLock()
	// defer s.mu.RUnlock()
	// value, ok := s.sessions[sessionID]
	// if !ok {
	// 	return Session{}, fmt.Errorf("session ID does not exist: %s", sessionID)
	// }
	// if value.Expiry.Before(time.Now()) { //session expired
	// 	return Session{}, fmt.Errorf("session expired")
	// }
	// return value, nil

	val, err := s.client.Get(context.Background(), sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return Session{}, fmt.Errorf("session ID does not exist: %s", sessionID)
		}
		return Session{}, fmt.Errorf("error getting session: %w", err)
	}
	var session Session
	err = json.Unmarshal([]byte(val), &session)
	if err != nil {
		return Session{}, fmt.Errorf("error unmarshalling session data: %w", err)
	}
	return session, nil
}

func (s *SessionStore) DeleteSession(sessionID string) error {
	// s.mu.Lock()
	// defer s.mu.Unlock()

	// _, exists := s.sessions[sessionID]
	// if !exists {
	// 	return fmt.Errorf("session not found")
	// }

	// delete(s.sessions, sessionID)
	result := s.client.Del(context.Background(), sessionID)
	if result.Err() != nil {
		return fmt.Errorf("error deleting session data: %w", result.Err())
	}
	if result.Val() == 0 {
		// No keys were deleted - session didn't exist
		return fmt.Errorf("session not found")
	}

	return nil
}

func NewSessionStore(redis *redis.Client) *SessionStore {
	return &SessionStore{
		client: redis,
	}
}
