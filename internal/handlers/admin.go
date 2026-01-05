package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/models"
	"github.com/kwachatracker/backend/internal/services"
)

type AdminHandler struct {
	FCMService      *services.FCMService
	GeminiService   *services.GeminiService
	InsightsHandler *InsightsHandler
}

// GetStats returns dashboard statistics
func (h *AdminHandler) GetStats(c *gin.Context) {
	var stats models.AdminStats

	// Total users
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Active users (synced in last 7 days)
	activeThreshold := time.Now().AddDate(0, 0, -7)
	database.DB.QueryRow(
		"SELECT COUNT(DISTINCT user_id) FROM transactions WHERE created_at >= $1",
		activeThreshold,
	).Scan(&stats.ActiveUsers7d)

	// Insights today
	todayStart := time.Now().Truncate(24 * time.Hour)
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM user_insights WHERE generated_at >= $1",
		todayStart,
	).Scan(&stats.InsightsToday)

	// Total transactions
	database.DB.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&stats.TotalTransactions)

	// Notifications sent today (if we track them)
	stats.NotificationsSentToday = 0 // TODO: implement when we add notifications table

	// API usage estimate
	stats.APIUsage.GeminiRequestsToday = stats.InsightsToday
	stats.APIUsage.EstimatedCost = float64(stats.InsightsToday) * 0.003 // rough estimate

	c.JSON(http.StatusOK, stats)
}

// GetUsers returns paginated user list
func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	filter := c.Query("filter") // "synced", "all"

	offset := (page - 1) * limit

	// Build query
	query := `
		SELECT 
			u.id, 
			u.device_id, 
			u.fcm_token,
			u.consent_given,
			u.created_at,
			MAX(t.created_at) as last_sync,
			COUNT(t.id) as transaction_count,
			COUNT(DISTINCT i.id) as insights_count
		FROM users u
		LEFT JOIN transactions t ON u.id = t.user_id
		LEFT JOIN user_insights i ON u.id = i.user_id
	`

	if filter == "synced" {
		query += " WHERE t.created_at >= $3"
	}

	query += " GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2"

	var rows *sql.Rows
	var err error

	if filter == "synced" {
		rows, err = database.DB.Query(query, limit, offset, time.Now().AddDate(0, 0, -7))
	} else {
		rows, err = database.DB.Query(query, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	users := []models.AdminUser{}
	for rows.Next() {
		var user models.AdminUser
		var lastSync sql.NullTime
		var fcmToken sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.DeviceID,
			&fcmToken,
			&user.ConsentAnalytics,
			&user.CreatedAt,
			&lastSync,
			&user.TransactionCount,
			&user.InsightsCount,
		)
		if err != nil {
			continue
		}

		if lastSync.Valid {
			user.LastSync = &lastSync.Time
		}
		if fcmToken.Valid {
			user.FCMToken = fcmToken.String
		}

		users = append(users, user)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM users"
	if filter == "synced" {
		countQuery += " WHERE EXISTS (SELECT 1 FROM transactions WHERE user_id = users.id AND created_at >= $1)"
		database.DB.QueryRow(countQuery, time.Now().AddDate(0, 0, -7)).Scan(&total)
	} else {
		database.DB.QueryRow(countQuery).Scan(&total)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
	})
}

// GetInsights returns paginated insights
func (h *AdminHandler) GetInsights(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	userID := c.Query("user_id")
	dateFrom := c.Query("date_from")

	offset := (page - 1) * limit

	// filters first
	query := `
		SELECT id, user_id, category, message, generated_at
		FROM user_insights
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if userID != "" {
		argCount++
		query += " AND user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
	}

	if dateFrom != "" {
		argCount++
		query += " AND generated_at >= $" + strconv.Itoa(argCount)
		args = append(args, dateFrom)
	}

	// Add LIMIT and OFFSET last
	argCount++
	query += " ORDER BY generated_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	query += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch insights", "details": err.Error()})
		return
	}
	defer rows.Close()

	insights := []models.AdminInsight{}
	for rows.Next() {
		var insight models.AdminInsight
		rows.Scan(
			&insight.ID,
			&insight.UserID,
			&insight.Type,
			&insight.Content,
			&insight.GeneratedAt,
		)
		// Default delivered to true for now since we don't track it per insight
		insight.Delivered = true
		insights = append(insights, insight)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM user_insights").Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"insights": insights,
		"total":    total,
		"page":     page,
	})
}

// TriggerInsights manually triggers AI analysis
func (h *AdminHandler) TriggerInsights(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}

	c.ShouldBindJSON(&req)

	// Note: Current implementation triggers for all users
	// TODO: Implement single-user analysis when needed
	go h.InsightsHandler.RunDailyAnalysis()

	if req.UserID != "" {
		c.JSON(http.StatusOK, gin.H{"message": "Analysis triggered (all users - single-user not yet implemented)"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Analysis triggered for all users"})
	}
}

// Broadcast sends push notifications
func (h *AdminHandler) Broadcast(c *gin.Context) {
	var req models.BroadcastRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tokens []string

	if req.Target == "specific" && len(req.UserIDs) > 0 {
		// Get tokens for specific users
		placeholders := ""
		for i := range req.UserIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "$" + strconv.Itoa(i+1)
		}
		query := "SELECT fcm_token FROM users WHERE id IN (" + placeholders + ") AND fcm_token IS NOT NULL"

		args := make([]interface{}, len(req.UserIDs))
		for i, id := range req.UserIDs {
			args[i] = id
		}

		rows, _ := database.DB.Query(query, args...)
		defer rows.Close()

		for rows.Next() {
			var token string
			rows.Scan(&token)
			tokens = append(tokens, token)
		}
	} else if req.Target == "active" {
		// Get tokens for active users (last 7 days)
		rows, _ := database.DB.Query(`
			SELECT DISTINCT u.fcm_token 
			FROM users u
			INNER JOIN transactions t ON u.id = t.user_id
			WHERE t.created_at >= $1 AND u.fcm_token IS NOT NULL
		`, time.Now().AddDate(0, 0, -7))
		defer rows.Close()

		for rows.Next() {
			var token string
			rows.Scan(&token)
			tokens = append(tokens, token)
		}
	} else {
		// Get all tokens
		rows, _ := database.DB.Query("SELECT fcm_token FROM users WHERE fcm_token IS NOT NULL")
		defer rows.Close()

		for rows.Next() {
			var token string
			rows.Scan(&token)
			tokens = append(tokens, token)
		}
	}

	if len(tokens) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tokens found"})
		return
	}

	// Send notifications (in background if scheduled)
	if req.ScheduledFor != nil && req.ScheduledFor.After(time.Now()) {
		// TODO: implement job queue for scheduled notifications
		c.JSON(http.StatusOK, gin.H{"message": "Scheduled notification (not yet implemented)"})
	} else {
		go func() {
			for _, token := range tokens {
				h.FCMService.SendNotification(c, token, req.Title, req.Body, nil)
			}
		}()
		c.JSON(http.StatusOK, gin.H{
			"message": "Broadcasting notification",
			"count":   len(tokens),
		})
	}
}

// GetTransactions returns paginated transactions
func (h *AdminHandler) GetTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	userID := c.Query("user_id")
	category := c.Query("category")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	offset := (page - 1) * limit

	query := "SELECT id, user_id, type, category, amount, balance, description, date FROM transactions WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if userID != "" {
		argCount++
		query += " AND user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
	}

	if category != "" {
		argCount++
		query += " AND category = $" + strconv.Itoa(argCount)
		args = append(args, category)
	}

	if dateFrom != "" {
		argCount++
		query += " AND date >= $" + strconv.Itoa(argCount)
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		argCount++
		query += " AND date <= $" + strconv.Itoa(argCount)
		args = append(args, dateTo)
	}

	// Add LIMIT and OFFSET last
	argCount++
	query += " ORDER BY date DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	query += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions", "details": err.Error()})
		return
	}
	defer rows.Close()

	transactions := []models.Transaction{}
	for rows.Next() {
		var txn models.Transaction
		var balance sql.NullFloat64
		var description sql.NullString

		rows.Scan(
			&txn.ID,
			&txn.UserID,
			&txn.Type,
			&txn.Category,
			&txn.Amount,
			&balance,
			&description,
			&txn.Date,
		)

		if balance.Valid {
			val := balance.Float64
			txn.Balance = &val
		}
		if description.Valid {
			txn.Description = &description.String
		}

		transactions = append(transactions, txn)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         page,
	})
}
