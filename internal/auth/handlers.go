package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthHandlers struct {
	Store      *store.UsersStore
	Sessions   *SessionStore
	Config     *oauth2.Config
	AdminEmail string
	JWTService *JWTService
}

func NewAuthHandlers(store *store.UsersStore, sessions *SessionStore, config *oauth2.Config, adminEmail string, jwtService *JWTService) *AuthHandlers {
	return &AuthHandlers{
		Store:      store,
		Sessions:   sessions,
		Config:     config,
		AdminEmail: adminEmail,
		JWTService: jwtService,
	}
}

func (h *AuthHandlers) GoogleLogin(c *gin.Context) {
	state := uuid.New().String()
	// Preserve popup parameter in state for callback
	if c.Query("popup") == "true" {
		state += "|popup=true"
	}
	c.SetCookie("state", state, 3600, "/", "localhost", false, true)
	redirect := h.Config.AuthCodeURL(state)
	c.Redirect(302, redirect)
}

func (h *AuthHandlers) GoogleCallback(c *gin.Context) {
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
	c.SetCookie("state", "", -1, "/", "", false, false)
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
	}
	isPopup := strings.Contains(stateCookie, "popup=true")

	if isPopup {
		// Return HTML for popup
		html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head><title>Authentication Successful</title></head>
    <body>
        <script>
            if (window.opener) {
                window.opener.postMessage({
                    token: "%s",
                    user: {
                        id: "%s",
                        email: "%s", 
                        role: "%s"
                    }
                }, "https://localhost:3000");  //
                window.close();
            } else {
                document.body.innerHTML = "<p>Authentication complete. You can close this window.</p>";
            }
        </script>
        <p>Authentication successful. This window should close automatically.</p>
    </body>
    </html>
`, JWTToken, user.ID, user.Email, user.Role)

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
