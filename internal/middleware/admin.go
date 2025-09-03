package middleware

import (
	"github.com/alexalbu001/iguanas-jewelry-api/internal/auth"
	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminMiddleware struct {
	Sessions *auth.SessionStore
	User     service.UsersStore
}

func (a *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			// RequireAuth should have set this!
			c.Error(&customerrors.ErrInternalServer)
			return
		}

		role, exists := c.Get("role")
		if !exists {
			// RequireAuth should have set this!
			c.Error(&customerrors.ErrInternalServer)
			return
		}

		if role == "admin" {
			c.Next()
			return
		}
		c.Error(&customerrors.ErrUserUnauthorized)
		return
	}
}

func NewAdminMiddleware(session *auth.SessionStore, userStore service.UsersStore) *AdminMiddleware {
	return &AdminMiddleware{
		Sessions: session,
		User:     userStore,
	}
}
