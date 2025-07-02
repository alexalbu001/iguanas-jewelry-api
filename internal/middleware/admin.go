package middleware

import (
	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/gin-gonic/gin"
)

type AdminMiddleware struct {
	Sessions *auth.SessionStore
	User     *store.UsersStore
}

func (a *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			// RequireAuth should have set this!
			c.AbortWithStatusJSON(500, gin.H{"error": "auth middleware failed"})
			return
		}

		user, err := a.User.GetUserByID(userID.(string))
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		if user.Role == "admin" {
			c.Set("role", user.Role)
			c.Next()
			return
		}
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}
}

func NewAdminMiddleware(session *auth.SessionStore, userStore *store.UsersStore) *AdminMiddleware {
	return &AdminMiddleware{
		Sessions: session,
		User:     userStore,
	}
}
