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
		reqToken, err := c.Cookie("jwt_token")

		if err != nil || reqToken == "" {
			// Fallback to Authorization header for backward compatibility
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.Error(&customerrors.ErrUserUnauthorized)
				return
			}
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				c.Error(&customerrors.ErrUserUnauthorized)
				return
			}
			reqToken = splitToken[1]
		}

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
