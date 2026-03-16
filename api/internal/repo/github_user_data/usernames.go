package github_data_repo

import "context"

func (r *GithubDataPsgrRepo) GetAllUsernames(ctx context.Context) ([]string, error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.GetAllUsernames"

	rows, err := r.db.QueryContext(ctx, `
	select username from github_data.users;
	`)

	if err != nil {
		return nil, r.handleError(err, fn+".selectUsernames")
	}

	usernames := []string{}

	for rows.Next() {
		username := ""
		err := rows.Scan(&username)
		if err != nil {
			r.logger.Error("unexcpected error occurred when scanning usernames", "source", fn, "err", err)
			return nil, r.handleError(err, fn+".scanUsername")
		}
		usernames = append(usernames, username)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("unexcpected error occurred after scanning usernames", "source", fn, "err", err)
		return nil, r.handleError(err, fn+".afterScanRowsError")
	}

	return usernames, nil
}
