package github_user_data

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/logger"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
	"github.com/jackc/pgx/v5/pgconn"
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

func TestStorageAddUserDataSucess(t *testing.T) {
	mock, repo := getMockAndRepo()
	githubRepo := domain.GithubRepository{ID: 123, OwnerUsername: "alex"}
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: []domain.GithubRepository{githubRepo},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`).WithArgs(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values (($1, $2, $3, $4, $5, $6, $7, $8));
	`).WithArgs(githubRepo.ID, githubRepo.OwnerUsername, githubRepo.PushedAt, githubRepo.UpdatedAt, githubRepo.Language, githubRepo.StarsCount, githubRepo.Fork, githubRepo.ForksCount).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.AddUserData(userData)
	require.NoError(t, err)
}

func TestStorageAddUserDataAlreadyExistsErrorCheck(t *testing.T) {
	mock, repo := getMockAndRepo()
	githubRepo := domain.GithubRepository{ID: 123, OwnerUsername: "alex"}
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: []domain.GithubRepository{githubRepo},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`).WithArgs(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt).WillReturnError(&pgconn.PgError{Code: "23505"})

	mock.ExpectRollback()

	require.Equal(t, repo.AddUserData(userData), &repoerr.ErrConflictValue{})
}

func TestStorageAddUserDataNoRepositoriesSuccess(t *testing.T) {
	mock, repo := getMockAndRepo()
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: make([]domain.GithubRepository, 0),
	}

	mock.ExpectBegin()

	mock.ExpectExec(`
	insert into users (username, name, company, location, bio, public_repos_count, followers_count, following_count, fetched_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`).WithArgs(userData.Username, userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	require.Equal(t, repo.AddUserData(userData), nil)
}

func TestStorageUpdateUserDataSuccess(t *testing.T) {
	mock, repo := getMockAndRepo()
	githubRepo := domain.GithubRepository{ID: 123, OwnerUsername: "alex"}
	userData := domain.GithubUserData{
		Username:     "alex",
		FetchedAt:    time.Now(),
		Repositories: []domain.GithubRepository{githubRepo},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`
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
	`).WithArgs(userData.Name, userData.Company, userData.Location, userData.Bio, userData.PublicRepos, userData.Followers, userData.Following, userData.FetchedAt, userData.Username).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
	insert into repositories (github_id, owner_username, pushed_at, updated_at, language, stars_count, is_fork, forks_count)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	on conflict (github_id)
	do update set
    pushed_at   = excluded.pushed_at,
    updated_at  = excluded.updated_at,
    language    = excluded.language,
    stars_count = excluded.stars_count,
    is_fork     = excluded.is_fork,
    forks_count = excluded.forks_count;
	`).WithArgs(githubRepo.ID, githubRepo.OwnerUsername, githubRepo.PushedAt, githubRepo.UpdatedAt, githubRepo.Language, githubRepo.StarsCount, githubRepo.Fork, githubRepo.ForksCount).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`
	delete from repositories r
	where r.owner_username = $1
	and not exists (
		select 1
		from (values ($2)) as v(github_id)
		where v.github_id = r.github_id
	);
	`).WithArgs(githubRepo.OwnerUsername, githubRepo.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.UpdateUserData(userData)
	require.NoError(t, err)
}
