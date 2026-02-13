# GitHub Banners API - Manual

## Prerequisites

- Go 1.25+
- PostgreSQL 14+
- GitHub Personal Access Token(s)

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# CORS allowed origins (comma-separated)
CORS_ORIGINS=example.com,www.example.com

# GitHub tokens for API access (comma-separated, supports multiple for rate limit distribution)
GITHUB_TOKENS=ghp_token1,ghp_token2

# Rate limiting
RATE_LIMIT_RPS=10

# Cache configuration (valid units: ms, s, m, h)
CACHE_TTL=5m

# Request timeout for external APIs
REQUEST_TIMEOUT=10s

# Logging (levels: DEBUG, INFO, WARN, ERROR)
LOG_LEVEL=DEBUG
LOG_FORMAT=json

# Secret key for inter-service communication
SERVICES_SECRET_KEY=your-secret-key

# PostgreSQL configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=banners
DB_HOST=localhost
PGPORT=5432
```

## Running the Service

### Local Development

```bash
# Install dependencies
go mod download

# Set environment variables (or use .env file with your preferred loader)
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=banners
export DB_HOST=localhost
export PGPORT=5432
export GITHUB_TOKENS=your_github_token

# Run the service
go run main.go
```

### Using Docker

```bash
# Build the image
docker build -t github-banners-api .

# Run the container
docker run -p 8080:8080 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=banners \
  -e DB_HOST=host.docker.internal \
  -e PGPORT=5432 \
  -e GITHUB_TOKENS=your_github_token \
  github-banners-api
```

### With Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: banners
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: banners
      DB_HOST: postgres
      PGPORT: "5432"
      GITHUB_TOKENS: your_token_here
      LOG_LEVEL: DEBUG
    depends_on:
      - postgres

volumes:
  postgres_data:
```

## Database Migrations

Migrations run automatically on startup using Goose. Manual commands:

```bash
# Install Goose CLI
go install github.com/pressly/goose/v3/cmd/goose@latest

# Apply migrations
goose postgres "postgres://user:pass@host:port/db" up

# Rollback last migration
goose postgres "postgres://user:pass@host:port/db" down

# Create new migration
goose -dir internal/migrations create migration_name sql
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Package Tests

```bash
# Test user stats service
go test ./internal/domain/user_stats/...

# Test preview usecase
go test ./internal/domain/preview/...

# Test repository
go test ./internal/repo/github_user_data/...

# Test renderer infrastructure
go test ./internal/infrastructure/renderer/...
```

### Run with Verbose Output

```bash
go test -v ./...
```

## API Usage

### Get Banner Preview

```bash
# Basic request
curl "http://localhost:8080/banners/preview/?username=torvalds&type=wide"

# Response: SVG image (Content-Type: image/svg+xml)
```

### Error Responses

```bash
# Invalid banner type
curl "http://localhost:8080/banners/preview/?username=torvalds&type=invalid"
# Response: {"error": "invalid banner type"}

# User not found
curl "http://localhost:8080/banners/preview/?username=nonexistentuser12345&type=wide"
# Response: {"error": "user doesn't exist"}

# Missing parameters
curl "http://localhost:8080/banners/preview/?username=torvalds"
# Response: {"error": "invalid inputs"}
```

## Monitoring & Observability

### Logs

The service outputs structured JSON logs when `LOG_FORMAT=json`:

```json
{
  "level": "info",
  "msg": "Starting github banners API service",
  "time": "2024-01-15T10:00:00Z"
}
```

Log levels:
- `DEBUG`: Detailed request/response info
- `INFO`: Service lifecycle events
- `WARN`: Non-critical errors
- `ERROR`: Critical errors

### Health Checks

Currently no dedicated health endpoint. Check service health via:

```bash
# Check if service is responding
curl -I http://localhost:8080/banners/preview/?username=test&type=wide
```

## Background Worker

The stats worker runs hourly by default and refreshes all cached user data:

- Interval: 1 hour (configurable in `main.go`)
- Batch size: 1 (configurable)
- Concurrency: 1 worker (configurable)

## Known Limitations

1. **POST /banners/** - Not implemented (returns 501)
2. **Kafka integration** - Code exists but not wired up in main.go
3. **Rate limiting** - Not currently enforced per-request
4. **Health endpoint** - Not implemented

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection
psql -h localhost -U postgres -d banners
```

### GitHub API Rate Limit

- Multiple tokens can be configured for rate limit distribution
- Check remaining quota: the fetcher logs token status on startup
- Tokens rotate automatically based on remaining quota

### Cache Issues

- Cache is in-memory only (lost on restart)
- TTL configurable via `CACHE_TTL`
- Clear by restarting the service

## Development

### Project Structure Convention

This project follows standard Go project layout:
- `cmd/` - Main applications (currently just `main.go` in root)
- `internal/` - Private application code
- `pkg/` - Public library code (none currently)

### Adding New Banner Types

1. Add type to `internal/domain/banner.go`:
   ```go
   const (
       TypeWide BannerType = iota
       TypeNarrow  // Add new type
   )
   ```

2. Update mappings:
   ```go
   var BannerTypes = map[string]BannerType{
       "wide":    TypeWide,
       "narrow":  TypeNarrow,
   }
   ```

3. Update renderer service to handle new type

### Adding New Endpoints

1. Create handler in `internal/handlers/`
2. Register route in `main.go`
3. Add to `docs/api.yaml`
