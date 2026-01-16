package domain

import (
	"time"

	"github.com/google/go-github/v81/github"
)

type UserStats struct {
	User         *github.User         `json:"user"`
	Repositories []*github.Repository `json:"repositories,omitempty"`
	Stats        Stats                `json:"stats"`
	FetchedAt    time.Time            `json:"fetched_at"`
	Cached       bool                 `json:"cached,omitempty"`
}

type Stats struct {
	TotalRepos    int            `json:"totalRepos"`
	OriginalRepos int            `json:"originalRepos"`
	ForkedRepos   int            `json:"forkedRepos"`
	TotalStars    int            `json:"totalStars"`
	TotalForks    int            `json:"totalForks"`
	Languages     map[string]int `json:"languages"`
}

// ErrorResponse for API error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type ServiceConfig struct {
	CacheTTL       time.Duration
	RequestTimeout time.Duration
}
