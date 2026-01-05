package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an app user
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	FCMToken     string    `json:"-" db:"fcm_token"`
	Operator     string    `json:"operator" db:"operator"`
	IsPremium    bool      `json:"is_premium" db:"is_premium"`
	ConsentGiven bool      `json:"consent_given" db:"consent_given"`
	ConsentDate  time.Time `json:"consent_date,omitempty" db:"consent_date"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction represents a mobile money transaction
type Transaction struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Type        string    `json:"type" db:"type"`         // INCOME, EXPENSE
	Category    string    `json:"category" db:"category"` // DATA, AIRTIME, PAYMENT, etc.
	Operator    string    `json:"operator" db:"operator"` // AIRTEL, MTN, ZAMTEL, ZEDMOBILE
	Recipient   *string   `json:"recipient,omitempty" db:"recipient"`
	Balance     *float64  `json:"balance,omitempty" db:"balance"`
	Reference   *string   `json:"reference,omitempty" db:"reference"`
	Description *string   `json:"description,omitempty" db:"description"`
	SMSHash     int       `json:"sms_hash" db:"sms_hash"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SyncRequest represents a batch of transactions to sync
type SyncRequest struct {
	DeviceID     string             `json:"device_id" binding:"required"`
	Transactions []TransactionInput `json:"transactions" binding:"required"`
	Timestamp    int64              `json:"timestamp" binding:"required"`
}

// TransactionInput represents incoming transaction data
type TransactionInput struct {
	Amount      float64  `json:"amount" binding:"required"`
	Type        string   `json:"type" binding:"required"`
	Category    string   `json:"category" binding:"required"`
	Operator    string   `json:"operator" binding:"required"`
	Recipient   *string  `json:"recipient,omitempty"`
	Balance     *float64 `json:"balance,omitempty"`
	Reference   *string  `json:"reference,omitempty"`
	Description *string  `json:"description,omitempty"`
	SMSHash     int      `json:"sms_hash" binding:"required"`
	Date        int64    `json:"date" binding:"required"` // Unix timestamp
}

// AnalyticsSummary represents spending analytics for a user
type AnalyticsSummary struct {
	TotalIncome      float64            `json:"total_income"`
	TotalExpenses    float64            `json:"total_expenses"`
	NetBalance       float64            `json:"net_balance"`
	ByCategory       map[string]float64 `json:"by_category"`
	ByOperator       map[string]float64 `json:"by_operator"`
	TransactionCount int                `json:"transaction_count"`
	Period           string             `json:"period"` // "week", "month", "all"
}

// PushNotification represents a notification to send
type PushNotification struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	ImageURL string            `json:"image_url,omitempty"`
}

// Admin Models

// AdminStats represents dashboard statistics
type AdminStats struct {
	TotalUsers             int      `json:"total_users"`
	ActiveUsers7d          int      `json:"active_users_7d"`
	InsightsToday          int      `json:"insights_today"`
	NotificationsSentToday int      `json:"notifications_sent_today"`
	TotalTransactions      int      `json:"total_transactions"`
	APIUsage               APIUsage `json:"api_usage"`
}

type APIUsage struct {
	GeminiRequestsToday int     `json:"gemini_requests_today"`
	EstimatedCost       float64 `json:"estimated_cost"`
}

// AdminUser represents user data for admin view
type AdminUser struct {
	ID               uuid.UUID  `json:"id"`
	DeviceID         string     `json:"device_id"`
	FCMToken         string     `json:"fcm_token"`
	ConsentAnalytics bool       `json:"consent_analytics"`
	ConsentAI        bool       `json:"consent_ai"`
	CreatedAt        time.Time  `json:"created_at"`
	LastSync         *time.Time `json:"last_sync,omitempty"`
	TransactionCount int        `json:"transaction_count"`
	InsightsCount    int        `json:"insights_count"`
}

// AdminInsight represents insight data for admin view
type AdminInsight struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Type           string    `json:"type"`
	Content        string    `json:"content"`
	GeneratedAt    time.Time `json:"generated_at"`
	Delivered      bool      `json:"delivered"`
	ResponseTimeMs int       `json:"response_time_ms,omitempty"`
}

// BroadcastRequest represents a push notification broadcast request
type BroadcastRequest struct {
	Title        string     `json:"title" binding:"required"`
	Body         string     `json:"body" binding:"required"`
	Target       string     `json:"target"` // "all", "active", "specific"
	UserIDs      []string   `json:"user_ids,omitempty"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
}
