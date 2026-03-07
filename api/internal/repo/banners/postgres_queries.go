package banners_repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
)

func (r *PostgresRepo) GetActiveBanners(ctx context.Context) ([]domain.LTBannerMetadata, error) {
	fn := "internal.repo.banners.PostgresRepo.GetActiveBanners"
	const q = `select github_username, banner_type, storage_path from banners where is_active = true`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		r.logger.Error("unexpected error when querying banners", "source", fn, "err", err)
		return nil, repoerr.ErrRepoInternal{Note: err.Error()}
	}

	defer rows.Close()

	res := make([]domain.LTBannerMetadata, 0)

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

		res = append(res, domain.LTBannerMetadata{
			Username:   username,
			BannerType: bt,
			UrlPath:    path,
			Active:     true,
		})
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("unexpected error after iterating banners", "source", fn, "err", err)
		return nil, repoerr.ErrRepoInternal{Note: err.Error()}
	}

	return res, nil
}

func (r *PostgresRepo) SaveBanner(ctx context.Context, b domain.LTBannerMetadata) error {
	fn := "internal.repo.banners.PostgresRepo.SaveBanner"
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

	const q = `
	insert into banners (github_username, banner_type, storage_path, is_active)
	values ($1, $2, $3, $4)
	on conflict (github_username, banner_type) do update set
		is_active = EXCLUDED.is_active,
		storage_path = EXCLUDED.storage_path;
	`

	_, err = r.db.ExecContext(ctx, q, b.Username, btStr, b.UrlPath, b.Active)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return repoerr.ErrConflictValue{Field: "github_username"}
		}
		r.logger.Error("unexpected error when upserting banner", "source", fn, "err", err)
		return repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return nil
}

func (r *PostgresRepo) DeactivateBanner(ctx context.Context, githubUsername string, bannerType domain.BannerType) error {
	fn := "internal.repo.banners.PostgresRepo.DeactivateBanner"
	if githubUsername == "" {
		return repoerr.ErrEmptyField{Field: "github_username"}
	}

	const q = `
	update banners
	set is_active = false
	where github_username = $1 and banner_type = $2 and is_active = true`

	res, err := r.db.ExecContext(ctx, q, githubUsername, domain.BannerTypesBackward[bannerType])
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

func (r *PostgresRepo) GetBanner(ctx context.Context, githubUsername string, bannerType domain.BannerType) (domain.LTBannerMetadata, error) {
	fn := "internal.repo.banners.PostgresRepo.GetBanner"
	const q = `
	select storage_path, is_active from banners
	where github_username = $1 and banner_type = $2;`
	meta := domain.LTBannerMetadata{Username: githubUsername, BannerType: bannerType}

	err := r.db.QueryRowContext(ctx, q, githubUsername, domain.BannerTypesBackward[bannerType]).Scan(&meta.UrlPath, &meta.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LTBannerMetadata{}, repoerr.ErrNothingFound
		}
		r.logger.Error("unexpected error when getting banner", "source", fn, "err", err)
		return domain.LTBannerMetadata{}, repoerr.ErrRepoInternal{Note: err.Error()}
	}
	return meta, nil
}
