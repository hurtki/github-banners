package domain

import (
	"time"
)

type GithubRepository struct {
	ID            *int64
	OwnerUsername string
	PushedAt      *time.Time
	UpdatedAt     *time.Time
	Language      *string
	StarsCount    int
	Fork          bool
	ForksCount    int
}

type GithubUserData struct {
	Username     *string
	Name         *string
	Company      *string
	Location     *string
	Bio          *string
	PublicRepos  *int
	Followers    *int
	Following    *int
	Repositories []GithubRepository
	FetchedAt    time.Time `json:"fetched_at"`
}

type GithubUserStats struct {
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
