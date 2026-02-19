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

| Feature                    | Description                                                                      |
| -------------------------- | -------------------------------------------------------------------------------- |
| **Statistics Aggregation** | Fetches and calculates total repositories, stars, forks, and language breakdowns |
| **Dynamic SVG Banners**    | Renders customizable banners with real-time user statistics                      |
| **Multi-Token Support**    | Supports multiple GitHub tokens for higher rate limits with automatic rotation   |
| **Smart Caching**          | Multi-layer caching (in-memory + PostgreSQL) with soft/hard TTL strategy         |
| **Background Refresh**     | Automatically refreshes statistics for tracked users                             |
| **Event-Driven**           | Kafka-based communication between microservices                                  |
| **Secure Communication**   | HMAC-based request signing for inter-service authentication                      |

---

## Tech Stack

<p align="center">
  <img src="https://skillicons.dev/icons?i=go,postgres,kafka,docker&theme=dark" alt="Tech Stack" />
</p>

| Component            | Technology                                           |
| -------------------- | ---------------------------------------------------- |
| **Language**         | Go 1.25.5                                            |
| **HTTP Router**      | [chi/v5](https://github.com/go-chi/chi)              |
| **Database**         | PostgreSQL 15                                        |
| **Message Queue**    | Apache Kafka ( Kraft manager )                       |
| **GitHub API**       | [go-github/v81](https://github.com/google/go-github) |
| **Migrations**       | [goose/v3](https://github.com/pressly/goose)         |
| **Caching**          | [go-cache](https://github.com/patrickmn/go-cache)    |
| **Containerization** | Docker / Docker Compose                              |

---

## Architecture

<img width="1679" height="856" alt="image" src="https://github.com/user-attachments/assets/dfb60c1c-c3f2-4ff2-bacb-f51008c8e9bd" />

### Github Stats Caching Strategy

| Layer          | TTL          | Behavior                                             |
| -------------- | ------------ | ---------------------------------------------------- |
| **Soft TTL**   | 10 minutes   | Serves cached data                                   |
| **Hard TTL**   | 24 hours     | Maximum cache lifetime , triggers background refresh |
| **In-Memory**  | Configurable | Fast access via go-cache                             |
| **PostgreSQL** | Persistent   | Durable storage for cached data                      |

### Database Schema

| Table          | Description                             |
| -------------- | --------------------------------------- |
| `banners`      | Banner configurations and storage paths |
| `users`        | GitHub user profile data                |
| `repositories` | Repository data linked to users         |

---

## Quick Start

### Requirements

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
# Service authentication
SERVICES_SECRET_KEY=your_secret_key
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

# PostgreSQL
POSTGRES_USER=github_banners
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=github_banners
DB_HOST=api-psgr
PGPORT=5432
```

`renderer/.env`:

```env
# DEBUG/INFO/WARN/ERROR
LOG_LEVEL=INFO
# text/json
LOG_FORMAT=json
# separated by comma list of broker instances
KAFKA_BROKERS_ADDRS=kafka:9092
```

**3. Start services**

```bash
docker-compose up --build
# Detached mode ( only build logs )
docker-compose up --build -d
```

### Development

```bash
# Run locally
cd api && go run main.go

# Run tests
./run_tests.sh
```

---

## Services

| Service    | Port             | Description              |
| ---------- | ---------------- | ------------------------ |
| `api`      | 80 ( public )    | Main API service         |
| `api-psgr` | 5432 ( private ) | PostgreSQL database      |
| `renderer` | -                | Banner rendering service |
| `storage`  | -                | Banner storage service   |
| `kafka`    | 9092 ( private ) | Apache Kafka broker      |

---

## API Endpoints

| Method | Endpoint                              | Description                                                    |
| ------ | ------------------------------------- | -------------------------------------------------------------- |
| `GET`  | `[api-service]/banners/preview`       | Get banner preview for a GitHub user ( not fully implemented ) |
| `POST` | `[api-service]/banners`               | Create a new banner ( not implemented )                        |
| `GET`  | `[storage-service]/{banner-url-path}` | Get long term banner ( not implemented )                       |

---

## Project Structure

```
github-banners/
├── api/                          # Main API service
│   ├── internal/
│   │   ├── app/user_stats/       # Background stats updating worker
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
│   │   └── repo/                 # Storages
│   │      ├── banners/           # long term banners storage
│   │      ├── github_user_data/  # github data
│   └── main.go
├── renderer/                     # Banner rendering service ( partially implemented )
│   ├── internal/
│   │   ├── infrastructure/       # External integrations
│   │   │   ├── kafka/            # kafka consumer group logic
│   │   ├── handlers/       # Evevnts handling and HTTP requests handling logic
│   └── main.go
├── storage/                      # Banner storage service ( not implemented )
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
