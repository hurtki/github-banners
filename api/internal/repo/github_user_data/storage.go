package github_user_data

import (
	"fmt"
	"strings"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/repo"
)

/*
Interface to implement
type GithubStatsRepo interface {
	AddUserData(domain.GithubUserData) error
	GetUserData(username string) (domain.GithubUserData, error)
	UpdateUserData(userData domain.GithubUserData) error
	GetAllUsernames() ([]string, error)
}
*/

func (r *GithubDataPsgrRepo) UpdateUserData(userData domain.GithubUserData) error {
	// update of the user table's row
	res, err := r.db.Exec(`
	update users
	set name = $1,
		company = $2,
		location = $3,
		bio = $4,
		public_repos_count = $5,
		followers_count = $6,
		following_count = $7,
		fetched_at = $8
	where username = $9;
	`, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt, userData.Username)

	if err != nil {
		return toRepoError(err)
	}

	// check if there were no rows affected, then nothing was updated -> ErrNothingChanged
	count, _ := res.RowsAffected()
	if count < 1 {
		return repo.ErrNothingChanged
	}

	// if a new data says that there is not repositories, then delete all existing ones
	if len(userData.Repositories) == 0 {
		_, err := r.db.Exec(`
		delete from repositories
		where owner_username = $1;
		`, userData.Username)
		return toRepoError(err)
	}

	// update of repositories
	// collect sql positional parameters($1, $2, $3...) for sql query
	// amd query args in order with positional parameters
	var (
		upsertPosParams []string
		upsertArgs      []any
	)

	i := 1
	const columnsCount = 8
	for _, item := range userData.Repositories {
		nums := make([]any, columnsCount)

		// generating numbers for postional parameters from i to i + columnsCount - 1
		// for example (1-8, 9-16, 17-4)
		for j := i; j < i+columnsCount; j++ {
			nums[j-i] = j
		}

		upsertPosParams = append(upsertPosParams, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", nums...))
		// collecting args in order with positional parameters
		upsertArgs = append(upsertArgs,
			item.ID,
			item.OwnerUsername,
			item.PushedAt,
			item.UpdatedAt,
			item.Language,
			item.StarsCount,
			item.Fork,
			item.ForksCount,
		)
		i += columnsCount
	}

	// now format the query, by inserting positional parameters into it. example: "($1, $2), ($3, $4), ($5, $6)"
	query := fmt.Sprintf(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values %s
	on conflict (github_id)
	do update set
    pushed_at   = excluded.pushed_at,
    updated_at  = excluded.updated_at,
    language    = excluded.language,
    stars_count = excluded.stars_count,
    is_fork     = excluded.is_fork,
    forks_count = excluded.forks_count;
	`, strings.Join(upsertPosParams, ", "),
	)

	// execute query with collected args
	_, err = r.db.Exec(query, upsertArgs...)

	if err != nil {
		return toRepoError(err)
	}

	deleteArgs := make([]any, len(userData.Repositories)+1)
	deleteArgs[0] = userData.Username
	var (
		deletePosParams []string
	)

	reposCount := len(userData.Repositories)

	// i here iterates on numbers of positional arguments
	// from 2 (because $1 is used for username in the query) to (repos count + 2)
	// but in deleteArgs we are going through indices, so we use i-1 ( to go from index 1)
	// but in deletePosParams we are on start so we use i-2 ( to go from index 0)
	for i := 2; i < reposCount+2; i++ {
		// fill of query args
		deleteArgs[i-1] = userData.Repositories[i-2].ID
		// fill of sql values
		deletePosParams[i-2] = fmt.Sprintf("($%d)", i)
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

	_, err = r.db.Exec(deleteQuery, deleteArgs...)

	return toRepoError(err)
}

func (r *GithubDataPsgrRepo) AddUserData(userData domain.GithubUserData) error {
	res, err := r.db.Exec(`
	insert into users (username, name, company, location, bio, publi_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`, userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt)

	if err != nil {
		return toRepoError(err)
	}

	count, _ := res.RowsAffected()
	if count < 1 {
		return repo.ErrNothingChanged
	}

}

func (r *GithubDataPsgrRepo) GetUserData(username string) (domain.GithubUserData, error) {
	row := r.db.QueryRow(`
	select (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at) from users
	where username = $1;
	`, username)

	data := domain.GithubUserData{}

	err := row.Scan(&data.Username, &data.Name, &data.Company, &data.Location, &data.Bio, &data.PublicRepos, &data.Followers, &data.Following, &data.FetchedAt)
	return data, toRepoError(err)
}
