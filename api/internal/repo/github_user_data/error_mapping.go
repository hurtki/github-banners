package github_user_data

import (
	"database/sql"
	"errors"

	repoerr "github.com/hurtki/github-banners/api/internal/repo"
	"github.com/jackc/pgx/v5/pgconn"
)

func toRepoError(err error) error {
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
		case "42601":
			return &repoerr.ErrRepoInternal{Note: pgErr.Hint}
		default:
			return &repoerr.ErrRepoInternal{Note: pgErr.Message}
		}
	}

	return &repoerr.ErrRepoInternal{Note: err.Error()}
}
