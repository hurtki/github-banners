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

- updated architecture pic

### Github Stats Caching Strategy

| Layer          | TTL          | Behavior                                             |
| -------------- | ------------ | ---------------------------------------------------- |
| **Soft TTL**   | 10 minutes   | Serves cached data                                   |
| **Hard TTL**   | 24 hours     | Maximum cache lifetime , triggers background refresh |
| **In-Memory**  | Configurable | Fast access via go-cache                             |
| **PostgreSQL** | Persistent   | Durable storage for cached data                      |

### Database Schema

| Table                      | Description                             |
| -------------------------- | --------------------------------------- |
| `banners`                  | Banner configurations and storage paths |
| `github_data.users`        | GitHub user profile data                |
| `github_data.repositories` | Repository data linked to users         |

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

> Use `.env.example` that lay in every folder

- Root `.env`
- `api/.env`
- `renderer/.env`
- `storage/.env`

**3. Start services**

> "dev" mode, 80 port without https

```bash
docker compose up --build
# Detached mode ( only build logs )
docker compose up --build -d
```

> "prod" mode, 443 port ( for cloudflare only )

Cloudflare configuration:
-pic

Where to get cert and key:
-pic

Put certificate `cert.pem` and private key `key.pem` to `/etc/nginx/ssl/`

### Development

> For testing consider using "dev" version of docker compose

```bash
# Run tests
./run_tests.sh

# CI static check:
# fix formatting
gofmt -s -w .
# tests check
./run_tests.sh
# spelling ( go install github.com/client9/misspell/cmd/misspell@latest )
# it will automatically fix all issues
misspell -source=auto -w .
# global api bundle ( only if you touched api.yaml files )
# It will combine description of global api in `/docs/api.yaml`
# into `bundled.yaml`
# for install https://redocly.com/docs/cli/installation
redocly bundle docs/api.yaml -o docs/bundled.yaml
```

---

## Services

| Service    | Port              | Description                  |
| ---------- | ----------------- | ---------------------------- |
| `nginx`    | 80/443 ( public ) | API gateway + static banners |
| `api`      | 80                | Main API service             |
| `api-psgr` | 5432              | PostgreSQL database          |
| `renderer` | 80                | Banner rendering service     |
| `storage`  | 80                | Banner storage service       |
| `kafka`    | 9092              | Apache Kafka broker          |

---

## API Endpoints

| Method | Endpoint             | Description                                                      |
| ------ | -------------------- | ---------------------------------------------------------------- |
| `GET`  | `/banners/preview`   | Get banner preview for a GitHub user                             |
| `POST` | `/banners`           | Create a new lont-term banner                                    |
| `GET`  | `/{banner-url-path}` | Get long term banner ( constantly updating since you created it) |

---

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Made with Go
</p>
