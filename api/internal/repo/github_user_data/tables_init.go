package github_user_data

import (
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/logger"
)

type GithubDataPsgrRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewGithubDataPsgrRepo(db *sql.DB, logger logger.Logger) (*GithubDataPsgrRepo, error) {
	repo := &GithubDataPsgrRepo{db: db, logger: logger}

	_, err := repo.db.Exec(`
create table if not exists users (
username     text primary key,
name         text ,
company      text,
location     text,
bio          text,
public_repos_count  int not null,
followers_count    int null null,
following_count    int null null,
fetched_at    timestamp not null
);

create table if not exists repositories (
github_id bigint primary key,
owner_username text not null,
pushed_at      timestamp,
updated_at     timestamp,
language      text,
stars_count    int not null,
is_fork          boolean not null,
forks_count    int not null,
constraint fk_repository_owner
  foreign key (owner_username)
  references users(username)
);
	`)

	if err != nil {
		return nil, err
	}

	return repo, nil
}
