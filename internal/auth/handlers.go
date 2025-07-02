package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandlers struct {
	Store    *store.UsersStore
	Sessions *SessionStore
}

func (h *AuthHandlers) GoogleLogin(c *gin.Context) {
	var state string
	state = uuid.New().String()
	c.SetCookie("state", state, 3600, "/", "localhost", false, true)
	redirect := conf.AuthCodeURL(state)
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

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "Could not exchange with google"})
		return
	}

	client := conf.Client(context.Background(), token)
	resp, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	c.SetCookie("state", "", -1, "/", "", false, false)
	defer resp.Body.Close()

	var userInfo googleInfo

	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "Could not parse"})
		return
	}

	user, err := h.Store.GetUserByGoogleID(userInfo.ID)
	if err != nil {
		newUser := models.User{
			GoogleID: userInfo.ID,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
		}
		user, err = h.Store.AddUser(newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create user"})
			return
		}
	}
	if user.Email == os.Getenv("ADMIN_EMAIL") && user.Role != "admin" {
		// Update user role in database
		err = h.Store.UpdateUserRole(user.ID, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create admin user"})
			return
		}
	}
	// Use sessions instead of this:
	// c.SetCookie("user_id", user.ID, 86400, "/", "", false, true)
	// c.SetCookie("user_email", user.Email, 86400, "/", "", false, true)
	sessionID := h.Sessions.CreateSession(user.ID, user.Email)
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

	c.Redirect(http.StatusFound, "/")
}

func NewAuthHandlers(store *store.UsersStore, sessions *SessionStore) *AuthHandlers {
	return &AuthHandlers{
		Store:    store,
		Sessions: sessions,
	}
}
