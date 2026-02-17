# GitHub Banners API - Manual

## Prerequisites

- Go 1.22+
- PostgreSQL 14+
- GitHub Personal Access Token(s)
- Renderer service (for banner rendering)
- Storage service (for persistent banner storage, optional)

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server port (default: 8080)
PORT=8080

# CORS allowed origins (comma-separated)
CORS_ORIGINS=example.com,www.example.com

# GitHub tokens for API access (comma-separated, supports multiple for rate limit distribution)
GITHUB_TOKENS=ghp_token1,ghp_token2

# Rate limiting (requests per second, currently not enforced)
RATE_LIMIT_RPS=10

# Cache configuration (valid units: ms, s, m, h)
CACHE_TTL=5m

# Request timeout for external APIs
REQUEST_TIMEOUT=10s

# Logging (levels: DEBUG, INFO, WARN, ERROR)
LOG_LEVEL=DEBUG
LOG_FORMAT=json

# Secret key for inter-service communication (HMAC signing)
SERVICES_SECRET_KEY=your-secret-key

# Renderer service URL
RENDERER_BASE_URL=https://renderer/

# Storage service URL
STORAGE_BASE_URL=http://storage/

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
export RENDERER_BASE_URL=http://localhost:3000/
export STORAGE_BASE_URL=http://localhost:3001/

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
  -e RENDERER_BASE_URL=http://renderer:3000/ \
  -e STORAGE_BASE_URL=http://storage:3001/ \
  -e SERVICES_SECRET_KEY=your-secret-key \
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
      RENDERER_BASE_URL: http://renderer:3000/
      STORAGE_BASE_URL: http://storage:3001/
      SERVICES_SECRET_KEY: your-secret-key
    depends_on:
      - postgres

  renderer:
    image: your-renderer-image
    ports:
      - "3000:3000"

  storage:
    image: your-storage-image
    ports:
      - "3001:3001"

volumes:
  postgres_data:
```

## Database Migrations

Migrations run automatically on startup using Goose (embedded SQL files). Manual commands:

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

### Migration Files

| File | Description |
|------|-------------|
| `001_create_users_table.sql` | Stores GitHub user profile data |
| `002_create_repositories_table.sql` | Stores user repositories data |
| `003_create_banners_table.sql` | Stores banner metadata (for future use) |

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

### Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `username` | Yes | GitHub username |
| `type` | Yes | Banner type (currently only "wide" supported) |

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

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success - returns SVG banner |
| 400 | Bad request - invalid parameters or banner type |
| 404 | Not found - user doesn't exist |
| 500 | Internal error - can't get preview (service unavailable) |
| 501 | Not implemented - POST /banners/ endpoint |

## Caching Strategy

### Stats Cache (User Statistics)

- **Storage**: In-memory (`patrickmn/go-cache`)
- **Soft TTL**: 10 minutes - data considered fresh
- **Hard TTL**: 24 hours - maximum cache lifetime
- **Behavior**:
  - Fresh data: returned immediately
  - Stale data: returned immediately, async refresh triggered
  - Cache miss: check DB, then GitHub API

### Preview Cache (Rendered Banners)

- **Storage**: In-memory with hash-based keys
- **Key**: Hash of BannerInfo (username + type + stats, excludes FetchedAt)
- **Purpose**: Deduplicate identical banner requests
- **Singleflight**: Prevents thundering herd on cache miss

## Background Worker

The stats worker runs periodically and refreshes cached user data:

- **Default interval**: 1 hour
- **Batch size**: 1 (configurable)
- **Concurrency**: 1 worker (configurable)
- **Process**: Fetches all usernames from DB, refreshes each from GitHub API

Configuration in `main.go`:
```go
statsWorker := user_stats_worker.NewStatsWorker(
    statsService.RefreshAll,
    time.Hour,  // interval
    logger,
    userstats.WorkerConfig{
        BatchSize: 1,
        Concurrency: 1,
        CacheTTL: time.Hour,
    },
)
```

## Known Limitations

1. **POST /banners/** - Not implemented (returns 501)
2. **Kafka integration** - Code exists but not wired up in main.go
3. **Rate limiting** - Not currently enforced per-request
4. **Health endpoint** - Not implemented
5. **Storage client** - Initialized but not used (for future banner persistence)

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
- Tokens rotate automatically based on remaining quota
- Check remaining quota via GitHub API: `curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/rate_limit`
- Fetcher logs token status on startup

### Renderer Service Issues

- Verify `RENDERER_BASE_URL` is correct and accessible
- Check `SERVICES_SECRET_KEY` matches between services
- Renderer must accept HMAC-signed requests

### Cache Issues

- Cache is in-memory only (lost on restart)
- TTL configurable via `CACHE_TTL`
- Clear by restarting the service

## Development

### Project Structure Convention

This project follows standard Go project layout:
- `main.go` - Application entry point
- `internal/` - Private application code
  - `domain/` - Business logic (entities, services)
  - `handlers/` - HTTP handlers
  - `infrastructure/` - External integrations
  - `repo/` - Data repositories
  - `cache/` - Caching layer
  - `config/` - Configuration

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

### Error Handling

Domain errors are defined in `internal/domain/errors.go`:
- `ErrNotFound` - Resource not found
- `ErrUnavailable` - Service unavailable
- `ConflictError` - Conflict with current state

Preview errors in `internal/domain/preview/errors.go`:
- `ErrInvalidBannerType`
- `ErrUserDoesntExist`
- `ErrInvalidInputs`
- `ErrCantGetPreview`