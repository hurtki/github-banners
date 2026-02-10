package banners

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
)

func (r *PostgresRepo) GetActiveBanners(ctx context.Context) ([]domain.LTBannerInfo, error) {
	fn := "internal.repo.banners.PostgresRepo.GetActiveBanners"
	const q = `SELECT github_username, banner_type, storage_path FROM banners WHERE is_active = true`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		r.logger.Error("unexpected error when querying banners", "source", fn, "err", err)
		return nil, repoerr.ErrRepoInternal{Note: err.Error()}
	}

	defer rows.Close()

	res := make([]domain.LTBannerInfo, 0)

	for rows.Next() {
		var username, btStr, path string
		if err := rows.Scan(&username, &btStr, &path); err != nil {
			r.logger.Error("unexpected error when scanning banners", "source", fn, "err", err)
			return nil, repoerr.ErrRepoInternal{Note: err.Error()}
		}

		bt, err := r.bannerTypeFromDB(btStr)
		if err != nil {
			return nil, err
		}

		res = append(res, domain.LTBannerInfo{
			BannerInfo: domain.BannerInfo{
				Username:   username,
				BannerType: bt,
			},
			UrlPath: path,
		})
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("unexpected error after iterating banners", "source", fn, "err", err)
		return nil, repoerr.ErrRepoInternal{Note: err.Error()}
	}

	return res, nil
}

func (r *PostgresRepo) AddBanner(ctx context.Context, b domain.LTBannerInfo) error {
	fn := "internal.repo.banners.PostgresRepo.AddBanner"
	if b.Username == "" {
		return repoerr.ErrEmptyField{Field: "github_username"}
	}
	if b.UrlPath == "" {
		return repoerr.ErrEmptyField{Field: "storage_path"}
	}

	btStr, err := r.bannerTypeToDB(b.BannerType)
	if err != nil {
		return err
	}

	const q = `insert into banners (github_username, banner_type, storage_path) values ($1, $2, $3)`

	_, err = r.db.ExecContext(ctx, q, b.Username, btStr, b.UrlPath)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return repoerr.ErrConflictValue{Field: "github_username"}
		}
		r.logger.Error("unexpected error when inserting banner", "source", fn, "err", err)
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return nil
}

func (r *PostgresRepo) DeactivateBanner(ctx context.Context, githubUsername string) error {
	fn := "internal.repo.banners.PostgresRepo.DeactivateBanner"
	if githubUsername == "" {
		return repoerr.ErrEmptyField{Field: "github_username"}
	}
	const q = `update banners set is_active = false where github_username = $1 and is_active = true`

	res, err := r.db.ExecContext(ctx, q, githubUsername)
	if err != nil {
		r.logger.Error("unexpected error when deactivating banner", "source", fn, "err", err)
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}

	affected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("unexpected error when reading affected rows", "source", fn, "err", err)
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}

	if affected == 0 {
		return repoerr.ErrNothingChanged
	}
	return nil
}

func (r *PostgresRepo) IsActive(ctx context.Context, githubUsername string) (bool, error) {
	fn := "internal.repo.banners.PostgresRepo.IsActive"
	const q = `select is_active from banners where github_username = $1`
	var active bool
	err := r.db.QueryRowContext(ctx, q, githubUsername).Scan(&active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, repoerr.ErrNothingFound
		}
		r.logger.Error("unexpected error when checking banner active state", "source", fn, "err", err)
		return false, repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return active, nil
}
