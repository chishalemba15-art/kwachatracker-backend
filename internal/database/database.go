package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB is the global database connection
var DB *sql.DB

// Connect initializes the database connection
func Connect(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("✅ Database connected successfully")
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			device_id VARCHAR(255) UNIQUE NOT NULL,
			fcm_token TEXT,
			operator VARCHAR(50) DEFAULT 'UNKNOWN',
			is_premium BOOLEAN DEFAULT FALSE,
			consent_given BOOLEAN DEFAULT FALSE,
			consent_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			amount DECIMAL(15, 2) NOT NULL,
			type VARCHAR(20) NOT NULL,
			category VARCHAR(50) NOT NULL,
			operator VARCHAR(50) NOT NULL,
			recipient TEXT,
			balance DECIMAL(15, 2),
			reference VARCHAR(100),
			description TEXT,
			sms_hash BIGINT NOT NULL,
			date TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, sms_hash)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_operator ON transactions(operator)`,

		// AI Insights table
		`CREATE TABLE IF NOT EXISTS user_insights (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			category VARCHAR(50) NOT NULL,
			priority VARCHAR(20) NOT NULL,
			generated_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_insights_user_id ON user_insights(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_insights_generated_at ON user_insights(generated_at)`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}
