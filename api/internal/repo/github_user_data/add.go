package github_user_data

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/repo"
)

func (r *GithubDataPsgrRepo) AddUserData(ctx context.Context, userData domain.GithubUserData) (err error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.AddUserData"
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
			r.logger.Error("error occured, when rolling back transaction", "err", rbErr, "source", fn)
		}
	}()

	// trying to insert user data
	_, err = tx.Exec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`, userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt)

	// if there is an error, return it converted to repo level error
	// including error with if userData with same username already exists, it will be converted to ErrConflictValue
	if err != nil {
		return toRepoError(err)
	}

	// Batch/Chunk Repository Upsert (Postgres Limit: 65535 parameters)
	batchSize := 500
	for i := 0; i < len(userData.Repositories); i += batchSize {
		end := min(i+batchSize, len(userData.Repositories))

		chunk := userData.Repositories[i:end]
		// inserting positional arguments into query
		// use of upsert, beacuse of edge case:
		/*
			user1 had repository and used our service
			user1 transfered repository to user2
			user2 started using our service before user1's info was update
			we are getting constraint error on insert because user1's repo is still in database and we are inserting with same github_id
		*/
		if err := r.upsertRepoBatch(ctx, tx, chunk); err != nil {
			return toRepoError(err)
		}
	}

	if err = tx.Commit(); err != nil {
		return toRepoError(err)
	}
	commited = true

	return nil
}
