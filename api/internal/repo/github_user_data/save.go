package github_data_repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/repo"
)

// UpdateUserData updates user's data (including his repositories) in database using transaction
func (r *GithubDataPsgrRepo) SaveUserData(ctx context.Context, userData domain.GithubUserData) (err error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.SaveUserData"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("can't start transaction", "source", fn, "err", err)
		return repo.ErrRepoInternal{
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

	_, err = tx.ExecContext(ctx, `
	insert into github_data.users (username, username_normalized, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, lower($2), $3, $4, $5, $6, $7, $8, $9, $10)
	on conflict (username_normalized) do update set
		username = EXCLUDED.username,
		name = EXCLUDED.name,
		company = EXCLUDED.company,
		location = EXCLUDED.location,
		bio = EXCLUDED.bio,
		public_repos_count = EXCLUDED.public_repos_count,
		followers_count = EXCLUDED.followers_count,
		following_count = EXCLUDED.following_count,
		fetched_at = EXCLUDED.fetched_at;
	`, userData.Username, userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt)
	if err != nil {
		return r.handleError(err, fn+".insertUser")
	}

	// if a new data says that there is no repositories, then delete all existing ones
	if len(userData.Repositories) == 0 {
		_, err := tx.ExecContext(ctx, `
		delete from github_data.repositories
		where owner_username_normalized = lower($1);
		`, userData.Username)

		if err != nil {
			return r.handleError(err, fn+".execDeleteAllRepositoriesFromUser")
		}

		err = tx.Commit()

		if err != nil {
			return r.handleError(err, fn+".commitAfterZeroRepositories")
		}
		committed = true
		return nil
	}

	// deduplicate
	seen := make(map[int64]struct{}, len(userData.Repositories))
	repos := make([]domain.GithubRepository, 0, len(userData.Repositories))

	for _, repo := range userData.Repositories {
		if _, ok := seen[repo.ID]; ok {
			continue
		}
		seen[repo.ID] = struct{}{}
		repos = append(repos, repo)
	}

	userData.Repositories = repos

	// Batch/Chunk Repository Upsert (Postgres positional parameters limit: 65535 parameters)
	batchSize := 500
	for i := 0; i < len(userData.Repositories); i += batchSize {
		end := min(i+batchSize, len(userData.Repositories))

		chunk := userData.Repositories[i:end]
		if err := r.upsertRepoBatch(ctx, tx, chunk); err != nil {
			return r.handleError(err, fn+".upsertRepoBatch")
		}
	}

	deleteArgs := make([]any, len(userData.Repositories)+1)
	deleteArgs[0] = userData.Username
	reposCount := len(userData.Repositories)

	deletePosParams := make([]string, reposCount)

	// i here iterates on numbers of positional arguments
	// from 2 (because $1 is used for username in the query) to (repos count + 2)
	// but in deleteArgs we are going through indices, so we use i-1 ( to go from index 1)
	// but in deletePosParams we are on start so we use i-2 ( to go from index 0)
	for i := 2; i < reposCount+2; i++ {
		// fill of query args
		deleteArgs[i-1] = userData.Repositories[i-2].ID
		// fill of sql values
		deletePosParams[i-2] = fmt.Sprintf("($%d::bigint)", i)
	}

	deleteQuery := fmt.Sprintf(`
		delete from github_data.repositories r
		where r.owner_username_normalized = lower($1)
		and not exists (
			select 1
			from (values %s) as v(github_id)
			where v.github_id = r.github_id
		);
	`, strings.Join(deletePosParams, ", "))

	if _, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...); err != nil {
		return r.handleError(err, fn+".deleteNotUsersRepositories")
	}

	if err = tx.Commit(); err != nil {
		return r.handleError(err, fn+".finalCommit")
	}
	committed = true
	return nil
}
