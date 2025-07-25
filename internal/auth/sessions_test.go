package auth_test

import (
	"log"
	"os"
	"testing"

	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var testSessionStore *auth.SessionStore

func TestMain(m *testing.M) {
	opt, err := redis.ParseURL(os.Getenv("TEST_REDIS_URL"))
	if err != nil {
		log.Fatal("Unable to connect to redis", err)
	}
	testRdb := redis.NewClient(opt)
	testSessionStore = auth.NewSessionStore(testRdb)
	defer testRdb.Close()
	os.Exit(m.Run())
}

func TestCreateSession(t *testing.T) {
	testUserId := uuid.New().String()
	testEmail := "test@gmail.com"

	sessionID, err := testSessionStore.CreateSession(testUserId, testEmail)
	if err != nil {
		t.Fatalf("error running CreateSession: %v", err)
	}
	defer testSessionStore.DeleteSession(sessionID)

	redisSession, err := testSessionStore.GetSession(sessionID)
	if err != nil {
		t.Fatalf("error running GetSession: %v", err)
	}

	if testUserId != redisSession.UserID {
		t.Errorf("expected %s userID, got %s instead", testUserId, redisSession.UserID)
	}
	if testEmail != redisSession.Email {
		t.Errorf("expected %s email, got %s instead", testEmail, redisSession.Email)
	}
}

func TestGetSession(t *testing.T) {
	_, err := testSessionStore.GetSession("not-a-real-id")
	if err == nil {
		t.Errorf("expected error and for non existent session, got nil")
	}

	_, err = testSessionStore.GetSession("")
	if err == nil {
		t.Errorf("expected error for empty session and got nil")
	}
}

func TestDeleteSession(t *testing.T) {
	testUserId := uuid.New().String()
	testEmail := "test@gmail.com"

	sessionID, err := testSessionStore.CreateSession(testUserId, testEmail)
	if err != nil {
		t.Fatalf("error running CreateSession: %v", err)
	}

	err = testSessionStore.DeleteSession(sessionID)
	if err != nil {
		t.Fatalf("error running DeleteSession: %v", err)
	}

	deletedSession, err := testSessionStore.GetSession(sessionID)
	if err == nil {
		t.Errorf("expected error for empty session retrieval and got nil")
	}
	var emptySession = auth.Session{}
	if deletedSession != emptySession {
		t.Errorf("expected nil session got: %v", deletedSession)
	}
}
