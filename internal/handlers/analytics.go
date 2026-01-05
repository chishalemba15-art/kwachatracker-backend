package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kwachatracker/backend/internal/database"
	"github.com/kwachatracker/backend/internal/models"
)

// AnalyticsHandler handles analytics endpoints
type AnalyticsHandler struct{}

// GetSummary returns spending analytics for the user
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	userID := c.GetString("user_id")
	period := c.DefaultQuery("period", "month")

	// Calculate date range
	var startDate time.Time
	now := time.Now()

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = time.Time{} // All time
	}

	summary := models.AnalyticsSummary{
		ByCategory: make(map[string]float64),
		ByOperator: make(map[string]float64),
		Period:     period,
	}

	// Get totals
	var totalIncome, totalExpenses sql.NullFloat64
	var count int

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) as expenses,
			COUNT(*) as count
		FROM transactions
		WHERE user_id = $1 AND date >= $2
	`

	err := database.DB.QueryRow(query, userID, startDate).Scan(&totalIncome, &totalExpenses, &count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate totals"})
		return
	}

	summary.TotalIncome = totalIncome.Float64
	summary.TotalExpenses = totalExpenses.Float64
	summary.NetBalance = summary.TotalIncome - summary.TotalExpenses
	summary.TransactionCount = count

	// Get breakdown by category
	categoryRows, err := database.DB.Query(`
		SELECT category, COALESCE(SUM(amount), 0) as total
		FROM transactions
		WHERE user_id = $1 AND type = 'EXPENSE' AND date >= $2
		GROUP BY category
		ORDER BY total DESC
	`, userID, startDate)

	if err == nil {
		defer categoryRows.Close()
		for categoryRows.Next() {
			var cat string
			var total float64
			if categoryRows.Scan(&cat, &total) == nil {
				summary.ByCategory[cat] = total
			}
		}
	}

	// Get breakdown by operator
	operatorRows, err := database.DB.Query(`
		SELECT operator, COALESCE(SUM(amount), 0) as total
		FROM transactions
		WHERE user_id = $1 AND date >= $2
		GROUP BY operator
		ORDER BY total DESC
	`, userID, startDate)

	if err == nil {
		defer operatorRows.Close()
		for operatorRows.Next() {
			var op string
			var total float64
			if operatorRows.Scan(&op, &total) == nil {
				summary.ByOperator[op] = total
			}
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetTrends returns spending trends over time
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	userID := c.GetString("user_id")
	period := c.DefaultQuery("period", "week") // daily grouping for week

	var startDate time.Time
	var groupFormat string
	now := time.Now()

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		groupFormat = "YYYY-MM-DD"
	case "month":
		startDate = now.AddDate(0, -1, 0)
		groupFormat = "YYYY-MM-DD"
	case "year":
		startDate = now.AddDate(-1, 0, 0)
		groupFormat = "YYYY-MM"
	default:
		startDate = now.AddDate(0, 0, -7)
		groupFormat = "YYYY-MM-DD"
	}

	rows, err := database.DB.Query(`
		SELECT 
			TO_CHAR(date, $3) as period,
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) as expenses
		FROM transactions
		WHERE user_id = $1 AND date >= $2
		GROUP BY TO_CHAR(date, $3)
		ORDER BY period ASC
	`, userID, startDate, groupFormat)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trends"})
		return
	}
	defer rows.Close()

	var trends []map[string]interface{}
	for rows.Next() {
		var p string
		var income, expenses float64
		if rows.Scan(&p, &income, &expenses) == nil {
			trends = append(trends, map[string]interface{}{
				"period":   p,
				"income":   income,
				"expenses": expenses,
				"net":      income - expenses,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"period": period,
	})
}
