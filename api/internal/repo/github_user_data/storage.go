package github_user_data

import (
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/logger"
)

type GithubDataPsgrRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewGithubDataPsgrRepo(db *sql.DB, logger logger.Logger) *GithubDataPsgrRepo {
	return &GithubDataPsgrRepo{db: db, logger: logger}
}
