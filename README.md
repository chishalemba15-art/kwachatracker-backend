# Kwacha Tracker Backend

A robust Go backend for the Kwacha Tracker mobile money tracking app.

## Features

- **JWT Authentication**: Secure device registration and token-based auth
- **Transaction Sync**: Batch sync with deduplication
- **Analytics**: Spending summaries and trends
- **Push Notifications**: Firebase Cloud Messaging integration
- **GDPR Compliance**: User data deletion endpoint

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis (optional, for rate limiting)
- Firebase service account (for push notifications)

### Setup

1. **Clone and install dependencies:**

```bash
cd backend
go mod tidy
```

1. **Set environment variables:**

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/kwachatracker?sslmode=disable"
export JWT_SECRET="your-secure-secret-key"
export PORT="8080"
export FIREBASE_CREDENTIALS="./firebase-credentials.json"
```

1. **Run the server:**

```bash
go run cmd/server/main.go
```

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/register` | Register device |

### Protected (requires Bearer token)

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/v1/consent` | Update consent status |
| DELETE | `/api/v1/data` | Delete all user data (GDPR) |
| POST | `/api/v1/sync` | Sync transactions |
| GET | `/api/v1/transactions` | Get transactions (paginated) |
| GET | `/api/v1/analytics/summary` | Spending summary |
| GET | `/api/v1/analytics/trends` | Spending trends |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `JWT_SECRET` | JWT signing secret | Required |
| `JWT_EXPIRATION_HOURS` | Token expiration | `720` (30 days) |
| `FIREBASE_CREDENTIALS` | Path to Firebase JSON | Optional |
| `ENVIRONMENT` | `development` or `production` | `development` |

## Deployment

### Docker

```bash
docker build -t kwachatracker-backend .
docker run -p 8080:8080 --env-file .env kwachatracker-backend
```

### Railway / Render / Fly.io

Deploy directly from Git with environment variables configured in the dashboard.
