# TaskDesk Backend Setup Walkthrough

I have successfully set up the Golang backend for TaskDesk using the Gin framework, following industry best practices and a clean directory structure. The backend is fully integrated with Supabase.

## Changes Made

### Project Structure

The project follows the standard Go layout:
- `cmd/api/main.go`: Entry point.
- `internal/`: Private code.
    - `api/handlers/`: Request handlers (`health.go`, `registration.go`).
    - `api/router/`: Route definitions (`router.go`).
    - `config/`: Configuration loader using Viper (`config.go`).
    - `db/`: Database connection pool using `pgxpool` (`db.go`).
    - `logger/`: Structured logging using Zap (`logger.go`).
    - `model/`: Data models (`registration.go`).
- `go.mod`, `go.sum`: Dependency management.
- `.env`: Environment variables (credentials included).

### Key Components

- **Gin Framework**: High-performance HTTP routing.
- **Supabase Integration**: Connected via the **Transaction Mode Pooler** for reliable IPv4/IPv6 support.
- **Viper & Zap**: Robust configuration management and structured logging.
- **Registration API**: A new endpoint `POST /api/v1/register` to save user data to the `registrations` table.

## Verification Results

### 1. Database Connectivity
The backend successfully connects to the Supabase database on startup.
```text
INFO    db/db.go:33     Successfully connected to Supabase Database!
```

### 2. Health & DB Status
The `/api/v1/health` endpoint verifies both the server and the database connection.
```bash
curl -i http://localhost:8080/api/v1/health
```
**Response:**
```json
{
  "db_status": "up",
  "status": "up"
}
```

### 3. Registration API Test
I verified the registration flow by creating a test user.
```bash
curl -i -X POST http://localhost:8080/api/v1/register \
-H "Content-Type: application/json" \
-d '{
  "full_name": "Ankit Raj",
  "email": "ankit@taskdesk.com",
  "organisation_name": "TaskDesk Team",
  "role": "Lead Backend Developer"
}'
```
**Response:**
```json
{
  "id": "395a775d-9ea0-4e6a-8fb3-09c94ad6b283",
  "full_name": "Ankit Raj",
  "email": "ankit@taskdesk.com",
  "organisation_name": "TaskDesk Team",
  "role": "Lead Backend Developer",
  "created_at": "2026-02-13T01:18:41.043143+05:30"
}
```

## How to Run

1. Ensure Go 1.22+ is installed.
2. The `.env` file is already configured with your Supabase credentials.
3. Start the server:
   ```bash
   go run cmd/api/main.go
   ```
4. The server runs on `http://localhost:8080`.
