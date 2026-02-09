package banners

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
)

func (r *PostgresRepo) GetActiveBanners(ctx context.Context) ([]domain.BannerMetadata, error) {
	const q = `SELECT github_username, banner_type, storage_path FROM banners WHERE is_active = true`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, repoerr.ErrRepoInternal{Note: err.Error()}
	}

	defer rows.Close()

	var res []domain.BannerMetadata

	for rows.Next() {
		var username, btStr, path string
		if err := rows.Scan(&username, &btStr, &path); err != nil {
			return nil, repoerr.ErrRepoInternal{Note: err.Error()}
		}

		bt, err := bannerTypeFromDB(btStr)
		if err != nil {
			return nil, err
		}

		res = append(res, domain.BannerMetadata{
			Username: username,
			BannerType: bt,
			UrlPath: path,
		})

	}

	if len(res) == 0 {
		return nil, repoerr.ErrNothingFound
	}

	return res, nil
}

func (r *PostgresRepo) AddBanner(ctx context.Context, b domain.LTBannerInfo) error {
	if b.Username == "" {
		return repoerr.ErrEmptyField{Field: "github_username"}
	}
	if b.UrlPath == "" {
		return repoerr.ErrEmptyField{Field: "storage_path"}
	}

	btStr, err := bannerTypeToDB(b.BannerType)
	if err != nil {
		return err
	}

	const q = `insert into banners (github_username, banner_type, storage_path) values ($1, $2, $3)`

	_, err = r.db.ExecContext(ctx, q, b.Username, btStr, b.UrlPath)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return repoerr.ErrConflictValue{}
		}
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return nil
}

func (r *PostgresRepo) DeactivateBanner(ctx context.Context, githubUsername string) error {
	if githubUsername == "" {
		return repoerr.ErrEmptyField{Field: "github_username"}
	}
	const q = `update banners set is_active = false where github_username = $1 and is_active = true`

	res, err := r.db.ExecContext(ctx, q, githubUsername)
	if err != nil {
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}

	if affected == 0 {
		return repoerr.ErrNothingFound
	}
	return nil
}

func (r *PostgresRepo) IsActive(ctx context.Context, githubUsername string) (bool, error) {
	const q = `select is_active from banners where github_username = $1`
	var active bool
	err := r.db.QueryRowContext(ctx, q, githubUsername).Scan(&active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, repoerr.ErrNothingFound
		}
		return false, repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return active, nil
}
