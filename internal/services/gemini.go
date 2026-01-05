package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GeminiService handles AI-powered transaction analysis
// Uses Gemini Developer API (REST) for lightweight integration
type GeminiService struct {
	apiKey     string
	httpClient *http.Client
	modelName  string
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiResponse represents the API response
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount int `json:"promptTokenCount"`
		TotalTokenCount  int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// SpendingData represents aggregated user spending for AI analysis
type SpendingData struct {
	UserID           string             `json:"user_id"`
	Period           string             `json:"period"` // "daily", "weekly", "monthly"
	TotalIncome      float64            `json:"total_income"`
	TotalExpenses    float64            `json:"total_expenses"`
	NetBalance       float64            `json:"net_balance"`
	ByCategory       map[string]float64 `json:"by_category"`
	TopMerchants     []string           `json:"top_merchants"`
	TransactionCount int                `json:"transaction_count"`
	SavingsDeposits  float64            `json:"savings_deposits"`
	PreviousPeriod   *SpendingData      `json:"previous_period,omitempty"`
}

// AIInsight represents generated insight for a user
type AIInsight struct {
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Category    string    `json:"category"` // "spending", "savings", "anomaly", "tip"
	Priority    string    `json:"priority"` // "high", "medium", "low"
	GeneratedAt time.Time `json:"generated_at"`
}

// NewGeminiService creates a new Gemini service
func NewGeminiService() (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	return &GeminiService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		modelName: "gemini-2.5-flash", // Fast and cost-effective
	}, nil
}

// AnalyzeSpending generates AI insights from spending data
func (s *GeminiService) AnalyzeSpending(ctx context.Context, data SpendingData) ([]AIInsight, error) {
	prompt := s.buildAnalysisPrompt(data)

	response, err := s.generateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	insights, err := s.parseInsights(response)
	if err != nil {
		// Fallback to basic insight if parsing fails
		log.Printf("Failed to parse AI response, using fallback: %v", err)
		return s.fallbackInsights(data), nil
	}

	return insights, nil
}

// buildAnalysisPrompt creates a structured prompt for spending analysis
func (s *GeminiService) buildAnalysisPrompt(data SpendingData) string {
	var categoryBreakdown strings.Builder
	for cat, amount := range data.ByCategory {
		categoryBreakdown.WriteString(fmt.Sprintf("- %s: K%.2f\n", cat, amount))
	}

	prompt := fmt.Sprintf(`You are a friendly financial advisor for a Zambian mobile money tracking app called "Kwacha Tracker".

Analyze this user's spending data and generate 2-3 personalized insights.

**Spending Data (%s):**
- Total Income: K%.2f
- Total Expenses: K%.2f
- Net Balance: K%.2f
- Savings Deposits: K%.2f
- Transaction Count: %d

**Category Breakdown:**
%s

**Instructions:**
1. Be encouraging and positive, especially about savings
2. Use Zambian Kwacha (K) for amounts
3. Keep each insight under 50 words
4. Focus on actionable tips
5. If savings > 10%% of income, congratulate them

**Output Format (JSON array):**
[
  {"title": "...", "message": "...", "category": "spending|savings|tip", "priority": "high|medium|low"}
]

Only output valid JSON, no additional text.`,
		data.Period,
		data.TotalIncome,
		data.TotalExpenses,
		data.NetBalance,
		data.SavingsDeposits,
		data.TransactionCount,
		categoryBreakdown.String(),
	)

	return prompt
}

// generateContent calls the Gemini API
func (s *GeminiService) generateContent(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.modelName,
		s.apiKey,
	)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{{Text: prompt}},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 500,
		},
		SafetySettings: []SafetySetting{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_ONLY_HIGH"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_ONLY_HIGH"},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// parseInsights extracts structured insights from AI response
func (s *GeminiService) parseInsights(response string) ([]AIInsight, error) {
	// Clean up response (remove markdown code blocks if present)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var rawInsights []struct {
		Title    string `json:"title"`
		Message  string `json:"message"`
		Category string `json:"category"`
		Priority string `json:"priority"`
	}

	if err := json.Unmarshal([]byte(response), &rawInsights); err != nil {
		return nil, fmt.Errorf("failed to unmarshal insights: %w", err)
	}

	insights := make([]AIInsight, len(rawInsights))
	for i, raw := range rawInsights {
		insights[i] = AIInsight{
			Title:       raw.Title,
			Message:     raw.Message,
			Category:    raw.Category,
			Priority:    raw.Priority,
			GeneratedAt: time.Now(),
		}
	}

	return insights, nil
}

// fallbackInsights returns pre-written insights when AI fails
func (s *GeminiService) fallbackInsights(data SpendingData) []AIInsight {
	insights := []AIInsight{}

	// Savings insight
	if data.SavingsDeposits > 0 {
		insights = append(insights, AIInsight{
			Title:       "💰 Great Saving Habit!",
			Message:     fmt.Sprintf("You've saved K%.0f this period. Keep it up!", data.SavingsDeposits),
			Category:    "savings",
			Priority:    "high",
			GeneratedAt: time.Now(),
		})
	}

	// Balance insight
	if data.NetBalance > 0 {
		insights = append(insights, AIInsight{
			Title:       "📈 Positive Balance",
			Message:     fmt.Sprintf("Your income exceeds expenses by K%.0f. Consider saving the surplus!", data.NetBalance),
			Category:    "tip",
			Priority:    "medium",
			GeneratedAt: time.Now(),
		})
	} else if data.NetBalance < 0 {
		insights = append(insights, AIInsight{
			Title:       "⚠️ Spending Alert",
			Message:     fmt.Sprintf("You've spent K%.0f more than earned. Review your expenses.", -data.NetBalance),
			Category:    "spending",
			Priority:    "high",
			GeneratedAt: time.Now(),
		})
	}

	return insights
}

// GenerateNotificationText creates a push notification from insights
func (s *GeminiService) GenerateNotificationText(insights []AIInsight) (title, body string) {
	if len(insights) == 0 {
		return "📊 Daily Summary", "Check your spending insights in the app!"
	}

	// Use highest priority insight for notification
	var best AIInsight
	for _, insight := range insights {
		if insight.Priority == "high" || best.Title == "" {
			best = insight
		}
	}

	return best.Title, best.Message
}
