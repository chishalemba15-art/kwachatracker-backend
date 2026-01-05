package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
// Supports both file path AND base64-encoded credentials from environment variable
func NewFCMService(credentialsPath string) (*FCMService, error) {
	ctx := context.Background()

	var opt option.ClientOption

	// First, check for base64-encoded credentials in environment variable
	if credentialsB64 := os.Getenv("FIREBASE_CREDENTIALS_BASE64"); credentialsB64 != "" {
		// Decode base64 credentials
		credentialsJSON, err := base64.StdEncoding.DecodeString(credentialsB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode FIREBASE_CREDENTIALS_BASE64: %w", err)
		}
		opt = option.WithCredentialsJSON(credentialsJSON)
		log.Println("📋 Using Firebase credentials from environment variable")
	} else if credentialsPath != "" {
		// Fall back to file path
		opt = option.WithCredentialsFile(credentialsPath)
		log.Println("📁 Using Firebase credentials from file")
	} else {
		return nil, fmt.Errorf("no Firebase credentials provided")
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Firebase Cloud Messaging initialized")
	return &FCMService{client: client}, nil
}

// SendNotification sends a push notification to a device
func (s *FCMService) SendNotification(ctx context.Context, token, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ClickAction: "OPEN_MAIN_ACTIVITY",
				ChannelID:   "kwachatracker_channel",
			},
		},
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		log.Printf("❌ Failed to send notification: %v", err)
		return err
	}

	log.Printf("✅ Notification sent: %s", response)
	return nil
}

// SendToMultiple sends notifications to multiple devices
func (s *FCMService) SendToMultiple(ctx context.Context, tokens []string, title, body string, data map[string]string) (int, int) {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		log.Printf("❌ Failed to send multicast: %v", err)
		return 0, len(tokens)
	}

	return response.SuccessCount, response.FailureCount
}

// SendWeeklySummary sends weekly summary notifications
func (s *FCMService) SendWeeklySummary(ctx context.Context, token string, income, expenses float64) error {
	net := income - expenses
	var emoji string
	if net >= 0 {
		emoji = "📈"
	} else {
		emoji = "📉"
	}

	title := fmt.Sprintf("%s Weekly Summary", emoji)
	body := fmt.Sprintf("Income: K%.0f | Expenses: K%.0f | Net: K%.0f", income, expenses, net)

	return s.SendNotification(ctx, token, title, body, map[string]string{
		"type": "weekly_summary",
	})
}

// SendBudgetAlert sends budget warning notifications
func (s *FCMService) SendBudgetAlert(ctx context.Context, token string, percentUsed int, budget float64) error {
	title := "⚠️ Budget Alert"
	body := fmt.Sprintf("You've used %d%% of your K%.0f monthly budget", percentUsed, budget)

	return s.SendNotification(ctx, token, title, body, map[string]string{
		"type": "budget_alert",
	})
}
