package github_data_repo

import (
	"database/sql"
	"errors"

	repoerr "github.com/hurtki/github-banners/api/internal/repo"
	"github.com/jackc/pgx/v5/pgconn"
)

// handleError converts error to repo errors and logs about
// important ones using fn as "source" field in logger to understand context of error
func (r *GithubDataPsgrRepo) handleError(err error, fn string) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return repoerr.ErrNothingFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return &repoerr.ErrConflictValue{Field: pgErr.ConstraintName}
		case "23502":
			return &repoerr.ErrEmptyField{Field: pgErr.ColumnName}
		// 42601 syntax error code for postgres
		case "42601":
			r.logger.Error("syntax error from postgres", "err", err, "source", fn)
			return &repoerr.ErrRepoInternal{Note: pgErr.Hint}
		default:
			r.logger.Error("unhandled PgError", "err", err, "source", fn)
			return &repoerr.ErrRepoInternal{Note: pgErr.Message}
		}
	}

	r.logger.Error("unhandled error from db", "err", err, "source", fn)
	return &repoerr.ErrRepoInternal{Note: err.Error()}
}
