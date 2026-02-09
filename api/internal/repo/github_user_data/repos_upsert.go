package github_user_data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
)

func (r *GithubDataPsgrRepo) upsertRepoBatch(ctx context.Context, tx *sql.Tx, batch []domain.GithubRepository) error {
	// before inserting into repositories, check if we are inserting at least one
	// if not skip
	if len(batch) < 1 {
		return nil
	}

	var (
		posParams []string
		args      []any
	)

	// building query for inserting repos
	i := 1
	for _, repo := range batch {
		tempPosArgs := []string{}
		for j := i; j < i+8; j++ {
			tempPosArgs = append(tempPosArgs, fmt.Sprintf("$%d", j))
		}
		posParams = append(posParams, fmt.Sprintf("(%s)", strings.Join(tempPosArgs, ", ")))
		args = append(args,
			repo.ID,
			repo.OwnerUsername,
			repo.PushedAt,
			repo.UpdatedAt,
			repo.Language,
			repo.StarsCount,
			repo.Fork,
			repo.ForksCount,
		)
		i += 8
	}

	query := fmt.Sprintf(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values %s
	on conflict (github_id) do update set
		owner_username = excluded.owner_username,
		pushed_at      = excluded.pushed_at,
		updated_at     = excluded.updated_at,
		language       = excluded.language,
		stars_count    = excluded.stars_count,
		is_fork        = excluded.is_fork,
		forks_count    = excluded.forks_count;
	`, strings.Join(posParams, ", "))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
