package github_user_data

import (
	"context"
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/domain"
)

func (r *GithubDataPsgrRepo) GetUserData(username string) (domain.GithubUserData, error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.GetUserData"
	// use of RepeatableRead/serializable sql isolation level to select user's data and his repos in same state
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

	row := tx.QueryRow(`
	select (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at) from users
	where username = $1;
	`, username)

	data := domain.GithubUserData{}

	err = row.Scan(&data.Username, &data.Name, &data.Company, &data.Location, &data.Bio, &data.PublicRepos, &data.Followers, &data.Following, &data.FetchedAt)

	rows, err := tx.Query(`
	select (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count) from repositories
	where owner_username = $1;
	`, username)

	if err != nil {
		return domain.GithubUserData{}, err
	}

	githubRepos := []domain.GithubRepository{}

	for rows.Next() {
		githubRepo := domain.GithubRepository{}
		err = rows.Scan(&githubRepo.ID, &githubRepo.OwnerUsername, &githubRepo.PushedAt, &githubRepo.UpdatedAt, &githubRepo.Language, &githubRepo.StarsCount, &githubRepo.Fork, &githubRepo.ForksCount)
		if err != nil {
			r.logger.Error("unexpected error from scan", "source", fn, "err", err)
			return domain.GithubUserData{}, toRepoError(err)
		}
		githubRepos = append(githubRepos, githubRepo)
	}

	// also check error after iterating rows
	// "Err returns the error, if any, that was encountered during iteration"
	err = rows.Err()
	if err != nil {
		r.logger.Error("unexpected error, after iterating rows", "source", fn, "err", err)
		return domain.GithubUserData{}, toRepoError(err)
	}

	data.Repositories = githubRepos

	return data, nil
}
