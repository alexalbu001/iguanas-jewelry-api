package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// generateCSRFToken generates a cryptographically secure CSRF token
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

type AuthHandlers struct {
	Store        *store.UsersStore
	Sessions     *SessionStore
	Config       *oauth2.Config
	AdminEmail   string
	AdminOrigin  string
	JWTService   *JWTService
	EmailService service.EmailService
	scheduler    gocron.Scheduler
}

func NewAuthHandlers(store *store.UsersStore, sessions *SessionStore, config *oauth2.Config, adminEmail, adminOrigin string, jwtService *JWTService, EmailService service.EmailService, scheduler gocron.Scheduler) *AuthHandlers {
	return &AuthHandlers{
		Store:        store,
		Sessions:     sessions,
		Config:       config,
		AdminEmail:   adminEmail,
		AdminOrigin:  adminOrigin,
		JWTService:   jwtService,
		EmailService: EmailService,
		scheduler:    scheduler,
	}
}

func (h *AuthHandlers) GoogleLogin(c *gin.Context) {
	state := uuid.New().String()
	// Preserve popup parameter and origin in state for callback
	if c.Query("popup") == "true" {
		origin := c.Query("origin")
		if origin == "" {
			// Default to admin origin if no origin specified
			origin = h.getAdminOrigin()
		}
		state += "|popup=true|origin=" + origin
	}

	// Use proper cookie settings for production
	domain, secure := h.getCookieSettings()
	c.SetCookie("state", state, 3600, "/", domain, secure, true)
	redirect := h.Config.AuthCodeURL(state)
	c.Redirect(302, redirect)
}

func (h *AuthHandlers) GoogleCallback(c *gin.Context) {
	logger, err := handlers.GetComponentLogger(c, "authentication")
	if err != nil {
		c.Error(err)
		return
	}
	stateGoogle := c.Query("state") // Gets from URL params
	stateCookie, err := c.Cookie("state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state cookie not found"})
		return
	}
	if stateGoogle != stateCookie {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No code"})
		return
	}

	token, err := h.Config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "Could not exchange with google"})
		return
	}

	client := h.Config.Client(c.Request.Context(), token)
	resp, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	// Clear state cookie with same domain settings
	domain, secure := h.getCookieSettings()
	c.SetCookie("state", "", -1, "/", domain, secure, true)
	defer resp.Body.Close()

	var userInfo googleInfo

	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "Could not parse"})
		return
	}

	user, err := h.Store.GetUserByGoogleID(c.Request.Context(), userInfo.ID)
	if err != nil {
		newUser := models.User{
			GoogleID: userInfo.ID,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
		}
		user, err = h.Store.AddUser(c.Request.Context(), newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user"})
			return
		}
		_, err = h.scheduler.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
			gocron.NewTask(h.EmailService.SendWelcome, context.Background(), user.Name, user.Email),
		)
		if err != nil {
			handlers.LogError(logger, "failed to schedule email job", err, "user_id", user.ID, "user_email", user.Email)
		}
	}
	if user.Email == h.AdminEmail && user.Role != "admin" {
		// Update user role in database
		err = h.Store.UpdateUserRole(c.Request.Context(), user.ID, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create admin user"})
			return
		}
	}
	// Use sessions instead of this:
	// c.SetCookie("user_id", user.ID, 86400, "/", "", false, true)
	// c.SetCookie("user_email", user.Email, 86400, "/", "", false, true)
	// sessionID, err := h.Sessions.CreateSession(user.ID, user.Email)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
	// 	return
	// }
	// c.SetCookie("session_id", sessionID, 86400, "/", "", true, true)
	JWTToken, err := h.JWTService.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.Error(err)
		return
	}

	// Generate CSRF token for additional security
	csrfToken := generateCSRFToken()

	// Set production-ready httpOnly cookies with security flags
	domain, secure = h.getCookieSettings()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt_token",
		Value:    JWTToken,
		MaxAge:   86400,
		Path:     "/",
		Domain:   domain, // Use correct domain from getCookieSettings
		Secure:   secure, // Use correct secure setting from getCookieSettings
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode, // Allows cross-site requests
	})

	// CSRF token - not httpOnly so frontend can read it for API requests
	c.SetCookie("csrf_token", csrfToken, 86400, "/", domain, secure, false)
	isPopup := strings.Contains(stateCookie, "popup=true")
	log.Printf("Setting cookies with domain='%s', secure=%t", domain, secure)
	log.Printf("AdminOrigin: %s", h.AdminOrigin)

	// Extract origin from state if it's a popup request
	var targetOrigin string
	if isPopup {
		// Parse origin from state cookie
		parts := strings.Split(stateCookie, "|")
		for _, part := range parts {
			if strings.HasPrefix(part, "origin=") {
				targetOrigin = strings.TrimPrefix(part, "origin=")
				break
			}
		}
		// Fallback to admin origin if not found
		if targetOrigin == "" {
			targetOrigin = h.getAdminOrigin()
		}
	}

	if isPopup {
		// Return minimal HTML for popup - just close immediately
		html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head><title>Authentication Successful</title></head>
    <body>
        <script>
            // Send message to parent window and close immediately
            if (window.opener && !window.opener.closed) {
                try {
                    window.opener.postMessage({
                        success: true,
                        user: {
                            id: "%s",
                            email: "%s", 
                            role: "%s"
                        },
                        csrfToken: "%s"
                    }, "%s");
                    
                    // Close immediately
                    window.close();
                } catch (error) {
                    console.error('Failed to send message to parent:', error);
                    // If we can't close, show minimal message
                    document.body.innerHTML = '<p>Authentication complete. You can close this window.</p>';
                }
            } else {
                document.body.innerHTML = '<p>Authentication complete. You can close this window.</p>';
            }
        </script>
    </body>
    </html>
`, user.ID, user.Email, user.Role, csrfToken, targetOrigin)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": JWTToken,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Logout handles user logout by clearing all httpOnly cookies
func (h *AuthHandlers) Logout(c *gin.Context) {
	// Clear all authentication cookies with same domain/path as set
	domain, secure := h.getCookieSettings()

	// Clear all authentication cookies with same security attributes
	c.SetCookie("jwt_token", "", -1, "/", domain, secure, true)
	c.SetCookie("user_id", "", -1, "/", domain, secure, true)
	c.SetCookie("user_email", "", -1, "/", domain, secure, true)
	c.SetCookie("user_role", "", -1, "/", domain, secure, true)
	c.SetCookie("csrf_token", "", -1, "/", domain, secure, false)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// setSecureCookie sets a cookie with all security attributes including SameSite
func (h *AuthHandlers) setSecureCookie(c *gin.Context, name, value string, maxAge int, path, domain string, secure, httpOnly bool, sameSite string) {
	// Use Go's built-in SetCookie method first
	c.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)

	// Then modify the Set-Cookie header to add SameSite attribute
	existingCookie := c.Writer.Header().Get("Set-Cookie")
	if existingCookie != "" {
		// Add SameSite to the existing cookie
		modifiedCookie := existingCookie + "; SameSite=" + sameSite
		c.Writer.Header().Set("Set-Cookie", modifiedCookie)
	}
}

// getCookieSettings returns the appropriate cookie domain and secure settings
// For localhost development, we need to set cookies that work for both frontend and admin
func (h *AuthHandlers) getCookieSettings() (domain string, secure bool) {
	// For localhost development, always use empty domain to allow cross-port access
	// and determine secure based on whether we're using HTTPS
	if strings.Contains(h.AdminOrigin, "localhost") {
		// Localhost development - use secure cookies if HTTPS, non-secure if HTTP
		secure = strings.HasPrefix(h.AdminOrigin, "https://")
		domain = "localhost" // Empty domain allows cross-port access on localhost
	} else {
		// Production environment - use secure cookies and proper domain
		secure = true
		// Extract domain from admin origin for production
		if h.AdminOrigin != "" {
			adminURL := strings.TrimPrefix(h.AdminOrigin, "https://")
			adminURL = strings.TrimPrefix(adminURL, "http://")
			if strings.Contains(adminURL, ".") {
				parts := strings.Split(adminURL, ".")
				if len(parts) >= 2 {
					// Use .domain.com to share cookies across all subdomains (www, api, etc.)
					domain = "." + strings.Join(parts[len(parts)-2:], ".") // .example.com
				}
			}
		}
	}
	return domain, secure
}

// getAdminOrigin returns the configured admin origin for secure postMessage
func (h *AuthHandlers) getAdminOrigin() string {
	if h.AdminOrigin != "" {
		return h.AdminOrigin
	}
	// Fallback for development
	return "http://localhost:3001"
}
