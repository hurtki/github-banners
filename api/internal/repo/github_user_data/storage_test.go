package github_user_data

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/logger"
	"github.com/stretchr/testify/require"
)

type LoggerMock struct{}

func (m LoggerMock) Debug(a string, b ...any)    {}
func (m LoggerMock) Info(a string, b ...any)     {}
func (m LoggerMock) Warn(a string, b ...any)     {}
func (m LoggerMock) Error(a string, b ...any)    {}
func (m LoggerMock) With(a ...any) logger.Logger { return m }

// helper for testing to create sql mock, create repoistory with it
func getMockAndRepo() (sqlmock.Sqlmock, *GithubDataPsgrRepo) {
	db, mock, _ := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	logger := LoggerMock{}

	repo := NewGithubDataPsgrRepo(db, logger)
	return mock, repo
}

func TestSaveUserDataSucess(t *testing.T) {
	mock, repo := getMockAndRepo()
	githubRepo1 := domain.GithubRepository{ID: 123, OwnerUsername: "alex"}
	githubRepo2 := domain.GithubRepository{ID: 45, OwnerUsername: "alex"}
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: []domain.GithubRepository{githubRepo1, githubRepo2},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	on conflict (username) do update set
		name = EXCLUDED.name,
		company = EXCLUDED.company,
		location = EXCLUDED.location,
		bio = EXCLUDED.bio,
		public_repos_count = EXCLUDED.public_repos_count,
		followers_count = EXCLUDED.followers_count,
		following_count = EXCLUDED.following_count,
		fetched_at = EXCLUDED.fetched_at;
	`).WithArgs(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values ($1, $2, $3, $4, $5, $6, $7, $8), ($9, $10, $11, $12, $13, $14, $15, $16)
	on conflict (github_id) do update set
		owner_username = excluded.owner_username,
		pushed_at      = excluded.pushed_at,
		updated_at     = excluded.updated_at,
		language       = excluded.language,
		stars_count    = excluded.stars_count,
		is_fork        = excluded.is_fork,
		forks_count    = excluded.forks_count;
	`).WithArgs(githubRepo1.ID, githubRepo1.OwnerUsername, githubRepo1.PushedAt, githubRepo1.UpdatedAt, githubRepo1.Language, githubRepo1.StarsCount, githubRepo1.Fork, githubRepo1.ForksCount, githubRepo2.ID, githubRepo2.OwnerUsername, githubRepo2.PushedAt, githubRepo2.UpdatedAt, githubRepo2.Language, githubRepo2.StarsCount, githubRepo2.Fork, githubRepo2.ForksCount).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
		delete from repositories r
		where r.owner_username = $1
		and not exists (
			select 1
			from (values ($2::bigint), ($3::bigint)) as v(github_id)
			where v.github_id = r.github_id
		);
	`).WithArgs(userData.Username, githubRepo1.ID, githubRepo2.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SaveUserData(context.TODO(), userData)
	require.NoError(t, err)
}

func TestSaveUserDataSucessNoRepos(t *testing.T) {
	mock, repo := getMockAndRepo()
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: []domain.GithubRepository{},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	on conflict (username) do update set
		name = EXCLUDED.name,
		company = EXCLUDED.company,
		location = EXCLUDED.location,
		bio = EXCLUDED.bio,
		public_repos_count = EXCLUDED.public_repos_count,
		followers_count = EXCLUDED.followers_count,
		following_count = EXCLUDED.following_count,
		fetched_at = EXCLUDED.fetched_at;
	`).WithArgs(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
		delete from repositories
		where owner_username = $1;
	`).WithArgs(userData.Username).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SaveUserData(context.TODO(), userData)
	require.NoError(t, err)
}

func TestGetAllUsernamesSuccess(t *testing.T) {
	mock, repo := getMockAndRepo()
	usernames := []string{"alex", "hurtki", "forge", "higor54"}

	usernameRows := sqlmock.NewRows([]string{"username"})
	for _, un := range usernames {
		usernameRows.AddRow(un)
	}
	mock.ExpectQuery(`
	select username from users;
	`).WillReturnRows(usernameRows)

	resUsernames, err := repo.GetAllUsernames(context.TODO())
	require.NoError(t, err)

	require.Equal(t, usernames, resUsernames)
}

func TestGetAllUsernamesNoUsernames(t *testing.T) {
	mock, repo := getMockAndRepo()

	mock.ExpectQuery(`
	select username from users;
	`).WillReturnRows(sqlmock.NewRows([]string{"username"}))

	resUsernames, err := repo.GetAllUsernames(context.TODO())
	require.NoError(t, err)

	require.Equal(t, []string{}, resUsernames)
}

func TestGetUserDataSuccess(t *testing.T) {
	mock, repo := getMockAndRepo()

	userData := domain.GithubUserData{Username: "Olivia"}
	repo1 := domain.GithubRepository{ID: 123, OwnerUsername: userData.Username}
	repo2 := domain.GithubRepository{ID: 3454, OwnerUsername: userData.Username}
	userData.Repositories = []domain.GithubRepository{repo1, repo2}
	userColumns := []string{"username", "name", "company", "location", "bio", "public_repos_count", "followers_count", "following_count", "fetched_at"}
	githubRepoColumns := []string{"github_id", "owner_username", "pushed_at", "updated_at", "language", "stars_count", "is_fork", "forks_count"}

	githubReposRows := sqlmock.NewRows(githubRepoColumns)
	for _, githubRepo := range userData.Repositories {
		githubReposRows.AddRow(githubRepo.ID, githubRepo.OwnerUsername, githubRepo.PushedAt, githubRepo.UpdatedAt, githubRepo.Language, githubRepo.StarsCount, githubRepo.Fork, githubRepo.ForksCount)
	}

	userRows := sqlmock.NewRows(userColumns)
	userRows.AddRow(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt)
	mock.ExpectBegin()

	mock.ExpectQuery(`
	select username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at from users
	where username = $1;
	`).WithArgs(userData.Username).WillReturnRows(userRows)

	mock.ExpectQuery(`
	select github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count from repositories
	where owner_username = $1;
	`).WithArgs(userData.Username).WillReturnRows(githubReposRows)
	mock.ExpectCommit()

	resUserData, err := repo.GetUserData(context.TODO(), userData.Username)
	require.NoError(t, err)
	require.Equal(t, userData, resUserData)
}
