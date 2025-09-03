package middleware

import (
	"strings"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/auth"
	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	JWTService *auth.JWTService
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// sessionID, err := c.Cookie("session_id")
		// if err != nil {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "no session cookie"})
		// 	return
		// }
		// retrievedSession, err := m.Sessions.GetSession(sessionID)
		// if err != nil {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		// 	return
		// }
		// if retrievedSession.Expiry.Before(time.Now()) {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "session expired"})
		// 	return
		// }
		// c.Set("userID", retrievedSession.UserID)
		// c.Set("userEmail", retrievedSession.Email)
		// c.Next()

		reqToken := c.GetHeader("Authorization")
		if reqToken == "" {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}
		reqToken = splitToken[1]

		claim, err := m.JWTService.ValidateToken(reqToken)
		if err != nil {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}

		if claim.Issuer != "iguanas-jewelry" {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}

		c.Set("userID", claim.UserID)
		c.Set("role", claim.Role)
		c.Next()
	}
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		JWTService: jwtService,
	}
}
