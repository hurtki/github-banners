package banners

import (
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/logger"
)

type PostgresRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewPostgresRepo(db *sql.DB, logger logger.Logger) *PostgresRepo {
	if logger != nil {
		logger = logger.With("repo", "banners")
	}
	return &PostgresRepo{
		db:     db,
		logger: logger,
	}
}
