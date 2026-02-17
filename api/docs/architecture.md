# Architecture Overview

## System Components

![alt text](image.png)

## Directory Structure

```
internal/
├── app/
│   └── user_stats/
│       └── worker.go          # Background worker for scheduled stats updates
├── cache/
│   ├── stats.go               # In-memory cache for user stats
│   ├── preview.go             # In-memory cache for rendered banners
│   └── banner_info_hash.go    # Hash utility for banner cache keys
├── config/
│   ├── config.go              # Main application config (env variables)
│   ├── kafka.go               # Kafka configuration (future use)
│   └── psgr.go                # PostgreSQL configuration
├── domain/
│   ├── banner.go              # Banner types, BannerInfo, LTBannerInfo, Banner structs
│   ├── types.go               # GithubRepository, GithubUserData, GithubUserStats models
│   ├── errors.go              # Domain errors (ErrNotFound, ErrUnavailable, ConflictError)
│   ├── preview/
│   │   ├── usecase.go         # PreviewUsecase - orchestrates stats + rendering
│   │   ├── service.go         # PreviewService - caches renderer results with singleflight
│   │   └── errors.go          # Preview-specific errors
│   └── user_stats/
│       ├── service.go         # UserStatsService - core business logic with cache strategy
│       ├── calculator.go      # CalculateStats - aggregates repository statistics
│       ├── models.go          # CachedStats, WorkerConfig structs
│       ├── cache.go           # Cache interface definition
│       └── interface.go       # Repository and fetcher interfaces
├── handlers/
│   ├── banners.go             # HTTP handlers for /banners/* endpoints
│   ├── dto.go                 # Future: DTOs for Create endpoint
│   └── error_response.go      # Error response helper
├── infrastructure/
│   ├── db/
│   │   └── connection.go      # PostgreSQL connection setup
│   ├── github/
│   │   ├── fetcher.go         # GitHub API client with rate limit handling
│   │   └── clients_pool.go    # Multi-token client pool for rate limit distribution
│   ├── kafka/
│   │   ├── producer.go        # Kafka event producer (future use)
│   │   └── dto.go             # Kafka event DTOs
│   ├── renderer/
│   │   ├── renderer.go        # Renderer HTTP client for banner rendering
│   │   ├── dto.go             # Renderer request/response DTOs
│   │   └── http/
│   │       └── client.go      # HTTP client factory for renderer
│   ├── httpauth/
│   │   ├── hmac_signer.go     # HMAC request signing for inter-service auth
│   │   └── round_tripper.go   # Auth HTTP round tripper
│   ├── server/
│   │   └── server.go          # HTTP server setup with CORS
│   └── storage/
│       ├── client.go          # Storage service HTTP client
│       └── dto.go            # Storage request/response DTOs
├── logger/
│   └── logger.go              # Structured logging (slog-based)
├── migrations/
│   ├── migrations.go          # Goose migration runner (embedded SQL files)
│   ├── 001_create_users_table.sql
│   ├── 002_create_repositories_table.sql
│   └── 003_create_banners_table.sql
└── repo/
    ├── banners/               # Banner repository (future use)
    │   ├── interface.go
    │   ├── postgres.go
    │   ├── postgres_mapper.go
    │   └── postgres_queries.go
    └── github_user_data/
        ├── storage.go         # GithubDataPsgrRepo struct
        ├── get.go             # GetUserData - fetch user from DB
        ├── save.go            # SaveUserData - persist user to DB
        ├── repos_upsert.go    # Batch upsert repositories
        ├── usernames.go       # GetAllUsernames - for worker refresh
        ├── error_mapping.go   # PostgreSQL error mapping
        └── storage_test.go    # Repository tests
```

## Data Flow

### Banner Preview Request

```
1. Client ──▶ GET /banners/preview/?username=X&type=wide

2. Handler (BannersHandler.Preview) ──▶ PreviewUsecase.GetPreview(username, type)

3. PreviewUsecase flow:
   ├─▶ Validate banner type
   ├─▶ StatsService.GetStats(username)
   │   └─▶ See StatsService flow below
   └─▶ PreviewProvider.GetPreview(bannerInfo)
       └─▶ See PreviewService flow below

4. StatsService.GetStats flow:
   ├─▶ Check in-memory cache
   │   ├─▶ If fresh (<10min): return cached stats
   │   └─▶ If stale (>10min but <24h): return cached, trigger async refresh
   ├─▶ If cache miss: Check database (repo.GetUserData)
   │   └─▶ If found: cache it, return stats
   └─▶ If db miss: Fetch from GitHub API (fetcher.FetchUserData)
       ├─▶ Save to database (repo.SaveUserData)
       ├─▶ Calculate stats (CalculateStats)
       ├─▶ Cache results
       └─▶ Return stats

5. PreviewService.GetPreview flow:
   ├─▶ Generate hash key from BannerInfo (excludes FetchedAt)
   ├─▶ Check in-memory cache by hash
   │   └─▶ If found: return cached banner
   └─▶ If miss: Render via singleflight (dedupe concurrent requests)
       ├─▶ Call renderer.RenderPreview(bannerInfo)
       │   └─▶ HTTP POST to renderer service (with HMAC auth)
       ├─▶ Cache result by hash
       └─▶ Return banner (SVG bytes)

6. Handler ──▶ Response (image/svg+xml)
```

### Background Worker Flow

```
StatsWorker.Start (runs every hour by default)
    │
    ▼
RefreshAll(ctx, config)
    │
    ├─▶ Get all usernames from database (repo.GetAllUsernames)
    │
    └─▶ Worker pool (configurable concurrency):
        └─▶ RecalculateAndSync(username)
            ├─▶ Fetch fresh data from GitHub API
            ├─▶ Save to database
            ├─▶ Calculate stats
            └─▶ Update cache
```

## Database Schema

### Users Table

```sql
CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY,
    name TEXT,
    company TEXT,
    location TEXT,
    bio TEXT,
    public_repos_count INT NOT NULL,
    followers_count INT,
    following_count INT,
    fetched_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_users_username ON users(username);
```

### Repositories Table

```sql
CREATE TABLE IF NOT EXISTS repositories (
    github_id BIGINT PRIMARY KEY, 
    owner_username TEXT NOT NULL, 
    pushed_at TIMESTAMP, 
    updated_at TIMESTAMP,
    language TEXT, 
    stars_count INT NOT NULL,
    forks_count INT NOT NULL,
    is_fork BOOLEAN NOT NULL,
    CONSTRAINT fk_repository_owner
        FOREIGN KEY (owner_username) 
        REFERENCES users(username) ON DELETE CASCADE
);

CREATE INDEX idx_repositories_owner_username ON repositories(owner_username);
```

### Banners Table

```sql
CREATE TABLE IF NOT EXISTS banners (
    github_username TEXT PRIMARY KEY, 
    banner_type TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_banners_github_username ON banners(github_username);
```

## Key Design Patterns

### 1. Clean Architecture / Hexagonal Architecture

- **Domain Layer**: Pure business logic (`domain/`, `domain/user_stats/`, `domain/preview/`)
- **Application Layer**: Use cases and app services (`app/`)
- **Infrastructure Layer**: External services (`infrastructure/`)
- **Interface Layer**: HTTP handlers (`handlers/`)

### 2. Repository Pattern

- `GithubUserDataRepository` interface for data persistence
- PostgreSQL implementation in `repo/github_user_data/`
- Abstracts database operations from domain logic

### 3. Cache-Aside Pattern with TTL Strategy

- **In-memory cache** checked first, then database, then external API
- **Two-tier TTL**:
  - Soft TTL (10 min): Data considered fresh, returned immediately
  - Hard TTL (24 hours): Data considered stale after soft TTL, returned but async refreshed
- **Preview cache**: Uses hash of BannerInfo (excluding FetchedAt) as key

### 4. Worker Pattern

- Background scheduled tasks via `StatsWorker`
- Concurrent processing with configurable batch size and worker count
- Results/errors collected via channels

### 5. Singleflight Pattern

- `PreviewService` uses singleflight to deduplicate concurrent rendering requests
- Same BannerInfo hash = same request, only one render call

### 6. Client Pool Pattern

- `clients_pool.go` manages multiple GitHub API tokens
- Automatic token rotation based on rate limit status
- Prevents single-token rate limit exhaustion

## External Dependencies

| Service    | Purpose                  | Library                |
| ---------- | ------------------------ | ---------------------- |
| PostgreSQL | Persistent storage       | `jackc/pgx/v5`         |
| GitHub API | User data source         | `google/go-github/v81` |
| Kafka      | Event streaming (future) | `IBM/sarama`           |
| Renderer   | Banner image generation  | HTTP client            |
| Storage    | Banner file storage      | HTTP client            |
| Goose      | Database migrations      | `pressly/goose/v3`     |
| Chi        | HTTP routing             | `go-chi/chi/v5`        |
| go-cache   | In-memory caching        | `patrickmn/go-cache`   |
| xxhash     | Fast hashing             | `cespare/xxhash/v2`    |
| singleflight | Request deduplication  | `golang.org/x/sync/singleflight` |

## Inter-Service Communication

Services communicate via HTTP with HMAC-based authentication:

```
API Service ──▶ HMAC Signer (signs request with timestamp + secret)
            ──▶ Auth Round Tripper (adds auth headers)
            ──▶ Renderer/Storage Service
```

Headers added:
- `X-Service`: Service identifier (e.g., "api")
- `X-Timestamp`: Unix timestamp
- `X-Signature`: HMAC-SHA256 of "service:timestamp"