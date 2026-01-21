package github_user_data

func (r *GithubDataPsgrRepo) AddUserData(userData domain.GithubUserData) error {
	fn := "internal.repo.github_user_data.GithubDataPsgrRepo.AddUserData"
	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("can't start transaction", "source", fn, "err", err)
		return repo.ErrRepoInternal{
			Note: err.Error(),
		}
	}

	defer func() {
		// if some error is being returned we will rollback transaction
		if err != nil {
			err := tx.Rollback()
			r.logger.Error("error, when rolling back transaction", "err", err, "source", fn)
		} else {
			err = tx.Commit()
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

	// before inserting into repositories, check if we are inserting at least one
	// if not skip
	if len(userData.Repositories) < 1 {
		return nil
	}

	var (
		posParams []string
		args      []any
	)

	// building query for inserting repos
	i := 1
	for _, repo := range userData.Repositories {
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
		i++
	}

	// inserting positional arguments into query
	query := fmt.Sprintf(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values (%s);
	`, strings.Join(posParams, ", "))

	_, err = tx.Exec(query, args...)

	return toRepoError(err)
}

