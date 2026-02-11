# GitHub Banners

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Apache Kafka](https://img.shields.io/badge/Apache%20Kafka-231F20?style=for-the-badge&logo=apachekafka&logoColor=white)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

---

A high-performance backend service that generates dynamic banners displaying GitHub user statistics. Perfect for enhancing your GitHub profile README with real-time stats.

## Overview

GitHub Banners fetches user data from the GitHub API, calculates aggregated statistics (repositories, stars, forks, languages), and renders beautiful SVG banners that automatically update.

### Key Features

| Feature | Description |
|---------|-------------|
| **Statistics Aggregation** | Fetches and calculates total repositories, stars, forks, and language breakdowns |
| **Dynamic SVG Banners** | Renders customizable banners with real-time user statistics |
| **Multi-Token Support** | Supports multiple GitHub tokens for higher rate limits with automatic rotation |
| **Smart Caching** | Multi-layer caching (in-memory + PostgreSQL) with soft/hard TTL strategy |
| **Background Refresh** | Automatically refreshes statistics for tracked users |
| **Event-Driven** | Kafka-based communication between microservices |
| **Secure Communication** | HMAC-based request signing for inter-service authentication |

---

## Tech Stack

<p align="center">
  <img src="https://skillicons.dev/icons?i=go,postgres,kafka,docker&theme=dark" alt="Tech Stack" />
</p>

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25 |
| **HTTP Router** | [chi/v5](https://github.com/go-chi/chi) |
| **Database** | PostgreSQL 15 |
| **Message Queue** | Apache Kafka |
| **GitHub API** | [go-github/v81](https://github.com/google/go-github) |
| **Migrations** | [goose/v3](https://github.com/pressly/goose) |
| **Caching** | [go-cache](https://github.com/patrickmn/go-cache) |
| **Containerization** | Docker / Docker Compose |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              GitHub Banners                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────┐     ┌─────────────┐     ┌──────────┐     ┌─────────────┐     │
│   │  Client │────▶│  API Service │────▶│  Kafka   │────▶│  Renderer   │     │
│   └─────────┘     └──────┬──────┘     └──────────┘     └──────┬──────┘     │
│                          │                                     │            │
│                          ▼                                     ▼            │
│                   ┌──────────────┐                      ┌─────────────┐     │
│                   │  PostgreSQL  │                      │   Storage   │     │
│                   └──────────────┘                      └─────────────┘     │
│                          ▲                                                  │
│                          │                                                  │
│                   ┌──────────────┐                                          │
│                   │ GitHub API   │                                          │
│                   │ (Multi-Token)│                                          │
│                   └──────────────┘                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Caching Strategy

| Layer | TTL | Behavior |
|-------|-----|----------|
| **Soft TTL** | 10 minutes | Serves cached data, triggers background refresh |
| **Hard TTL** | 24 hours | Maximum cache lifetime |
| **In-Memory** | Configurable | Fast access via go-cache |
| **PostgreSQL** | Persistent | Durable storage for cached data |

### Database Schema

| Table | Description |
|-------|-------------|
| `users` | GitHub user profile data |
| `repositories` | Repository data linked to users |
| `banners` | Banner configurations and storage paths |

---

## Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- GitHub Personal Access Token(s)

### Installation

**1. Clone the repository**

```bash
git clone https://github.com/yourusername/github-banners.git
cd github-banners
```

**2. Configure environment variables**

Root `.env`:
```env
API_INTERNAL_SERVER_PORT=80
```

`api/.env`:
```env
# CORS
CORS_ORIGINS=example.com,www.example.com

# GitHub tokens (comma-separated for load balancing)
GITHUB_TOKENS=ghp_token1,ghp_token2

# Rate limiting & Cache
RATE_LIMIT_RPS=10
CACHE_TTL=5m
REQUEST_TIMEOUT=10s

# Logging
LOG_LEVEL=DEBUG
LOG_FORMAT=json

# Service authentication
SERVICES_SECRET_KEY=your_secret_key

# PostgreSQL
POSTGRES_USER=github_banners
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=github_banners
DB_HOST=api-psgr
PGPORT=5432
```

**3. Start services**

```bash
docker-compose up --build
```

### Development

```bash
# Run locally
cd api && go run main.go

# Run tests
./run_tests.sh
```

---

## Configuration Reference

### Application Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |
| `GITHUB_TOKENS` | - | GitHub API tokens (comma-separated) |
| `RATE_LIMIT_RPS` | `10` | Rate limit (requests/second) |
| `CACHE_TTL` | `5m` | Default cache TTL |
| `REQUEST_TIMEOUT` | `10s` | HTTP request timeout |
| `LOG_LEVEL` | `info` | debug / info / warn / error |
| `LOG_FORMAT` | `json` | json / text |
| `SERVICES_SECRET_KEY` | `1234` | HMAC signing key |

### Database Settings

| Variable | Description |
|----------|-------------|
| `POSTGRES_USER` | Database user |
| `POSTGRES_PASSWORD` | Database password |
| `POSTGRES_DB` | Database name |
| `DB_HOST` | Database host |
| `PGPORT` | Database port |

---

## Services

| Service | Port | Description |
|---------|------|-------------|
| `api` | 80 | Main API service |
| `api-psgr` | 5432 | PostgreSQL database |
| `renderer` | - | Banner rendering service |
| `storage` | - | Banner storage service |
| `kafka` | 9092 | Apache Kafka broker |
| `zookeeper` | 2181 | Kafka coordination |

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/banners/preview` | Get banner preview for a GitHub user |
| `POST` | `/banners` | Create a new banner |

---

## Project Structure

```
github-banners/
├── api/                          # Main API service
│   ├── internal/
│   │   ├── app/user_stats/       # Background worker
│   │   ├── cache/                # In-memory cache
│   │   ├── config/               # Configuration
│   │   ├── domain/               # Business logic
│   │   │   ├── preview/          # Banner preview use case
│   │   │   └── user_stats/       # Statistics service
│   │   ├── handlers/             # HTTP handlers
│   │   ├── infrastructure/       # External integrations
│   │   │   ├── db/               # Database connection
│   │   │   ├── github/           # GitHub API client pool
│   │   │   ├── kafka/            # Kafka producer
│   │   │   ├── renderer/         # Renderer client
│   │   │   └── server/           # HTTP server
│   │   ├── migrations/           # SQL migrations
│   │   └── repo/                 # Data repositories
│   └── main.go
├── renderer/                     # Banner rendering service
├── storage/                      # Banner storage service
├── docker-compose.yaml
└── run_tests.sh
```

---

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Made with Go
</p>
