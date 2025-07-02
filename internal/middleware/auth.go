package middleware

import (
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	Sessions *auth.SessionStore
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "no session cookie"})
			return
		}
		retrievedSession, err := m.Sessions.GetSession(sessionID)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		if retrievedSession.Expiry.Before(time.Now()) {
			c.AbortWithStatusJSON(401, gin.H{"error": "session expired"})
			return
		}
		c.Set("userID", retrievedSession.UserID)
		c.Set("userEmail", retrievedSession.Email)
		c.Next()
	}
}

func NewAuthMiddleware(s *auth.SessionStore) *AuthMiddleware {
	return &AuthMiddleware{
		Sessions: s,
	}
}
