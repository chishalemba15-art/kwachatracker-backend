package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/models"
)

// SyncHandler handles transaction synchronization
type SyncHandler struct{}

// Sync receives and stores transactions from the app
func (h *SyncHandler) Sync(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify consent before syncing
	var consentGiven bool
	err := database.DB.QueryRow(
		"SELECT consent_given FROM users WHERE id = $1",
		userID,
	).Scan(&consentGiven)

	if err != nil || !consentGiven {
		c.JSON(http.StatusForbidden, gin.H{"error": "User consent required before syncing data"})
		return
	}

	// Begin transaction for batch insert
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	insertedCount := 0
	skippedCount := 0

	for _, t := range req.Transactions {
		// Use UPSERT to handle duplicates gracefully
		result, err := tx.Exec(`
			INSERT INTO transactions (id, user_id, amount, type, category, operator, recipient, balance, reference, description, sms_hash, date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (user_id, sms_hash) DO NOTHING
		`,
			uuid.New(),
			userID,
			t.Amount,
			t.Type,
			t.Category,
			t.Operator,
			t.Recipient,
			t.Balance,
			t.Reference,
			t.Description,
			t.SMSHash,
			time.UnixMilli(t.Date),
		)

		if err != nil {
			// Log but continue with other transactions
			skippedCount++
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			insertedCount++
		} else {
			skippedCount++ // Duplicate
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Sync completed",
		"inserted": insertedCount,
		"skipped":  skippedCount,
		"total":    len(req.Transactions),
	})
}

// GetTransactions retrieves user's transactions with pagination
func (h *SyncHandler) GetTransactions(c *gin.Context) {
	userID := c.GetString("user_id")

	// Pagination
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := parseInt(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	rows, err := database.DB.Query(`
		SELECT id, amount, type, category, operator, recipient, balance, reference, description, date
		FROM transactions
		WHERE user_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var t struct {
			ID          uuid.UUID
			Amount      float64
			Type        string
			Category    string
			Operator    string
			Recipient   *string
			Balance     *float64
			Reference   *string
			Description *string
			Date        time.Time
		}

		if err := rows.Scan(&t.ID, &t.Amount, &t.Type, &t.Category, &t.Operator,
			&t.Recipient, &t.Balance, &t.Reference, &t.Description, &t.Date); err != nil {
			continue
		}

		transactions = append(transactions, map[string]interface{}{
			"id":          t.ID,
			"amount":      t.Amount,
			"type":        t.Type,
			"category":    t.Category,
			"operator":    t.Operator,
			"recipient":   t.Recipient,
			"balance":     t.Balance,
			"reference":   t.Reference,
			"description": t.Description,
			"date":        t.Date.UnixMilli(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
