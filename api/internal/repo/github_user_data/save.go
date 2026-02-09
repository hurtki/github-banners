package github_user_data

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
	commited := false
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if !commited {
			rbErr := tx.Rollback()
			if rbErr != nil {
				r.logger.Error("error occured, when rolling back transaction", "err", rbErr, "source", fn)
			}
		}
	}()

	_, err = tx.ExecContext(ctx, `
	INSERT INTO users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (username) DO UPDATE SET
		name = EXCLUDED.name,
		company = EXCLUDED.company,
		location = EXCLUDED.location,
		bio = EXCLUDED.bio,
		public_repos_count = EXCLUDED.public_repos_count,
		followers_count = EXCLUDED.followers_count,
		following_count = EXCLUDED.following_count,
		fetched_at = EXCLUDED.fetched_at;
	`, userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt)
	if err != nil {
		return toRepoError(err)
	}

	// if a new data says that there is not repositories, then delete all existing ones
	if len(userData.Repositories) == 0 {
		_, err := tx.Exec(`
		delete from repositories
		where owner_username = $1;
		`, userData.Username)
		return toRepoError(err)
	}

	// Batch/Chunk Repository Upsert (Postgres Limit: 65535 parameters)
	batchSize := 500
	for i := 0; i < len(userData.Repositories); i += batchSize {
		end := min(i+batchSize, len(userData.Repositories))

		chunk := userData.Repositories[i:end]
		if err := r.upsertRepoBatch(ctx, tx, chunk); err != nil {
			return toRepoError(err)
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
		delete from repositories r
		where r.owner_username = $1
		and not exists (
			select 1
			from (values %s) as v(github_id)
			where v.github_id = r.github_id
		);
	`, strings.Join(deletePosParams, ", "))

	if _, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...); err != nil {
		return toRepoError(err)
	}

	if err = tx.Commit(); err != nil {
		return toRepoError(err)
	}
	commited = true
	return nil
}
