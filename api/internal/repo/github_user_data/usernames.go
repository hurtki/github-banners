package github_user_data

func (r *GithubDataPsgrRepo) GetAllUsernames() ([]string, error) {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.GetAllUsernames"

	rows, err := r.db.Query(`
	select username from users;
	`)

	if err != nil {
		return nil, toRepoError(err)
	}

	usernames := []string{}

	for rows.Next() {
		username := ""
		err := rows.Scan(&username)
		if err != nil {
			r.logger.Error("unexcpected error occured when scanning usernames", "source", fn, "err", err)
			return nil, toRepoError(err)
		}
		usernames = append(usernames, username)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("unexcpected error occured after scanning usernames", "source", fn, "err", err)
		return nil, toRepoError(err)
	}

	return usernames, nil
}
