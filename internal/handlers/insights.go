package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/services"
)

// InsightsHandler handles AI-powered insights endpoints
type InsightsHandler struct {
	gemini *services.GeminiService
	fcm    *services.FCMService
}

// NewInsightsHandler creates a new insights handler
func NewInsightsHandler(gemini *services.GeminiService, fcm *services.FCMService) *InsightsHandler {
	return &InsightsHandler{
		gemini: gemini,
		fcm:    fcm,
	}
}

// GenerateInsights generates AI insights for a specific user
func (h *InsightsHandler) GenerateInsights(c *gin.Context) {
	userID := c.GetString("user_id")

	// Fetch spending data for the last 24 hours
	spendingData, err := h.fetchSpendingData(userID, "daily")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spending data"})
		return
	}

	// Skip if no transactions
	if spendingData.TransactionCount == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":  "No transactions to analyze",
			"insights": []services.AIInsight{},
		})
		return
	}

	// Generate AI insights
	insights, err := h.gemini.AnalyzeSpending(c.Request.Context(), *spendingData)
	if err != nil {
		log.Printf("AI analysis failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"insights": insights,
		"period":   "daily",
		"analyzed": spendingData.TransactionCount,
	})
}

// RunDailyAnalysis processes all users with consent - called by scheduler
func (h *InsightsHandler) RunDailyAnalysis() {
	log.Println("🔄 Starting daily AI analysis job...")

	// Get all users with consent and FCM tokens
	rows, err := database.DB.Query(`
		SELECT id, fcm_token 
		FROM users 
		WHERE consent_given = true AND fcm_token IS NOT NULL
	`)
	if err != nil {
		log.Printf("❌ Failed to fetch users: %v", err)
		return
	}
	defer rows.Close()

	successCount := 0
	errorCount := 0

	for rows.Next() {
		var userID string
		var fcmToken sql.NullString

		if err := rows.Scan(&userID, &fcmToken); err != nil {
			continue
		}

		// Fetch user's spending data
		spendingData, err := h.fetchSpendingData(userID, "daily")
		if err != nil || spendingData.TransactionCount == 0 {
			continue
		}

		// Generate AI insights
		insights, err := h.gemini.AnalyzeSpending(nil, *spendingData)
		if err != nil {
			log.Printf("❌ AI analysis failed for user %s: %v", userID, err)
			errorCount++
			continue
		}

		// Store insights for retrieval
		h.storeInsights(userID, insights)

		// Send push notification if user has FCM token
		if fcmToken.Valid && h.fcm != nil {
			title, body := h.gemini.GenerateNotificationText(insights)
			err = h.fcm.SendNotification(nil, fcmToken.String, title, body, map[string]string{
				"type": "daily_insight",
			})
			if err != nil {
				log.Printf("⚠️ Push failed for user %s: %v", userID, err)
			}
		}

		successCount++

		// Rate limit to avoid overwhelming APIs
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("✅ Daily analysis complete: %d success, %d errors", successCount, errorCount)
}

// fetchSpendingData retrieves aggregated spending for a user
func (h *InsightsHandler) fetchSpendingData(userID, period string) (*services.SpendingData, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -1)
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
	case "monthly":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(0, 0, -1)
	}

	data := &services.SpendingData{
		UserID:     userID,
		Period:     period,
		ByCategory: make(map[string]float64),
	}

	// Get totals
	var totalIncome, totalExpenses sql.NullFloat64
	err := database.DB.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0),
			COUNT(*)
		FROM transactions
		WHERE user_id = $1 AND date >= $2
	`, userID, startDate).Scan(&totalIncome, &totalExpenses, &data.TransactionCount)

	if err != nil {
		return nil, err
	}

	data.TotalIncome = totalIncome.Float64
	data.TotalExpenses = totalExpenses.Float64
	data.NetBalance = data.TotalIncome - data.TotalExpenses

	// Get category breakdown
	rows, err := database.DB.Query(`
		SELECT category, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND type = 'EXPENSE' AND date >= $2
		GROUP BY category
	`, userID, startDate)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cat string
			var amount float64
			if rows.Scan(&cat, &amount) == nil {
				data.ByCategory[cat] = amount
			}
		}
	}

	// Get savings deposits
	var savingsDeposits sql.NullFloat64
	database.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND category = 'SAVINGS' AND type = 'EXPENSE' AND date >= $2
	`, userID, startDate).Scan(&savingsDeposits)
	data.SavingsDeposits = savingsDeposits.Float64

	return data, nil
}

// storeInsights saves generated insights to database
func (h *InsightsHandler) storeInsights(userID string, insights []services.AIInsight) {
	for _, insight := range insights {
		database.DB.Exec(`
			INSERT INTO user_insights (user_id, title, message, category, priority, generated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, insight.Title, insight.Message, insight.Category, insight.Priority, insight.GeneratedAt)
	}
}

// GetUserInsights retrieves stored insights for display
func (h *InsightsHandler) GetUserInsights(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := database.DB.Query(`
		SELECT title, message, category, priority, generated_at
		FROM user_insights
		WHERE user_id = $1
		ORDER BY generated_at DESC
		LIMIT 10
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch insights"})
		return
	}
	defer rows.Close()

	var insights []services.AIInsight
	for rows.Next() {
		var insight services.AIInsight
		if rows.Scan(&insight.Title, &insight.Message, &insight.Category, &insight.Priority, &insight.GeneratedAt) == nil {
			insights = append(insights, insight)
		}
	}

	c.JSON(http.StatusOK, gin.H{"insights": insights})
}
