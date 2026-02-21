package github_data_repo

import (
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/logger"
)

type GithubDataPsgrRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewGithubDataPsgrRepo(db *sql.DB, logger logger.Logger) *GithubDataPsgrRepo {
	return &GithubDataPsgrRepo{db: db, logger: logger.With("service", "github-data-repo")}
}
