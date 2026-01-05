package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kwachatracker/backend/config"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	Config *config.Config
}

// RegisterRequest represents a device registration request
type RegisterRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	FCMToken string `json:"fcm_token,omitempty"`
	Operator string `json:"operator,omitempty"`
}

// Register registers a new device or returns existing token
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var userID uuid.UUID
	var exists bool

	err := database.DB.QueryRow(
		"SELECT id FROM users WHERE device_id = $1",
		req.DeviceID,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		// Create new user
		userID = uuid.New()
		_, err = database.DB.Exec(
			`INSERT INTO users (id, device_id, fcm_token, operator) VALUES ($1, $2, $3, $4)`,
			userID, req.DeviceID, req.FCMToken, req.Operator,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		exists = false
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		exists = true
		// Update FCM token if provided
		if req.FCMToken != "" {
			database.DB.Exec("UPDATE users SET fcm_token = $1, updated_at = $2 WHERE id = $3",
				req.FCMToken, time.Now(), userID)
		}
	}

	// Generate JWT
	token, err := middleware.GenerateToken(
		userID.String(),
		req.DeviceID,
		h.Config.JWTSecret,
		h.Config.JWTExpiration,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID.String(),
		"token":       token,
		"is_new_user": !exists,
		"expires_in":  h.Config.JWTExpiration * 3600,
	})
}

// UpdateConsent updates the user's consent status
func (h *AuthHandler) UpdateConsent(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		ConsentGiven bool `json:"consent_given"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var consentDate *time.Time
	if req.ConsentGiven {
		now := time.Now()
		consentDate = &now
	}

	_, err := database.DB.Exec(
		`UPDATE users SET consent_given = $1, consent_date = $2, updated_at = $3 WHERE id = $4`,
		req.ConsentGiven, consentDate, time.Now(), userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent updated"})
}

// DeleteData deletes all user data (GDPR compliance)
func (h *AuthHandler) DeleteData(c *gin.Context) {
	userID := c.GetString("user_id")

	// Delete transactions first (foreign key)
	_, err := database.DB.Exec("DELETE FROM transactions WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transactions"})
		return
	}

	// Delete user
	_, err = database.DB.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All data deleted"})
}
