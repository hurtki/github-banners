package github_data_repo

import (
	"context"
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/repo"
)

func (r *GithubDataPsgrRepo) GetUserData(ctx context.Context, username string) (domain.GithubUserData, error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.GetUserData"
	// use of RepeatableRead/serializable sql isolation level to select user's data and his repos in same state
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		r.logger.Error("can't start transaction", "source", fn, "err", err)
		return domain.GithubUserData{}, repo.ErrRepoInternal{
			Note: err.Error(),
		}
	}
	committed := false
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if !committed {
			rbErr := tx.Rollback()
			if rbErr != nil {
				r.logger.Error("error occured, when rolling back transaction", "err", rbErr, "source", fn)
			}
		}
	}()

	row := tx.QueryRowContext(ctx, `
	select username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at from github_data.users
	where username_normalized = lower($1);
	`, username)

	data := domain.GithubUserData{}

	err = row.Scan(&data.Username, &data.Name, &data.Company, &data.Location, &data.Bio, &data.PublicRepos, &data.Followers, &data.Following, &data.FetchedAt)

	if err != nil {
		return domain.GithubUserData{}, r.handleError(err, fn+".scanIntoGithubUserData")
	}

	rows, err := tx.QueryContext(ctx, `
	select github_id, pushed_at, updated_at, language, stars_count, is_fork, forks_count from github_data.repositories
	where owner_username_normalized = lower($1);
	`, username)

	if err != nil {
		return domain.GithubUserData{}, r.handleError(err, fn+".selectRepositoriesQuery")
	}

	githubRepos := []domain.GithubRepository{}

	for rows.Next() {
		githubRepo := domain.GithubRepository{}
		err = rows.Scan(&githubRepo.ID, &githubRepo.PushedAt, &githubRepo.UpdatedAt, &githubRepo.Language, &githubRepo.StarsCount, &githubRepo.Fork, &githubRepo.ForksCount)
		if err != nil {
			return domain.GithubUserData{}, r.handleError(err, fn+".scanRepositoryRow")
		}
		githubRepo.OwnerUsername = data.Username
		githubRepos = append(githubRepos, githubRepo)
	}

	// also check error after iterating rows
	// "Err returns the error, if any, that was encountered during iteration"
	err = rows.Err()
	if err != nil {
		r.logger.Error("unexpected error, after iterating rows", "source", fn, "err", err)
		return domain.GithubUserData{}, r.handleError(err, fn+".afterIteratingRowsError")
	}

	data.Repositories = githubRepos

	if err = tx.Commit(); err != nil {
		return domain.GithubUserData{}, r.handleError(err, fn+".commit")
	}
	committed = true
	return data, nil
}
