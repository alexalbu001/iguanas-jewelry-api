package middleware

import (
	"strings"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/gin-gonic/gin"
)

func ValidateCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation for GET requests (they should be safe)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/webhooks/") {
			c.Next()
			return
		}

		// Get CSRF token from cookie
		csrfCookie, err := c.Cookie("csrf_token")
		if err != nil {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}

		// Get CSRF token from header
		csrfHeader := c.GetHeader("X-CSRF-Token")
		if csrfHeader == "" {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}

		// Validate tokens match
		if csrfCookie != csrfHeader {
			c.Error(&customerrors.ErrUserUnauthorized)
			return
		}

		c.Next()
	}
}
