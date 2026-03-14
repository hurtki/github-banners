package github_data_repo

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
			// don't forget to use lower for OwnerUsername for normalization
			// 1 (2) 3 4 5 6 7 8
			// 9 (10) 11 12 13 14
			// ...
			if j%8 == 2 {
				tempPosArgs = append(tempPosArgs, fmt.Sprintf("lower($%d)", j))
			} else {
				tempPosArgs = append(tempPosArgs, fmt.Sprintf("$%d", j))
			}
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
	insert into github_data.repositories (github_id, owner_username_normalized, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values %s
	on conflict (github_id) do update set
		owner_username_normalized = excluded.owner_username_normalized,
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
