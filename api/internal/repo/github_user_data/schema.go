package github_user_data

import "database/sql"

// Intialzies table for GithubDataPsgrRepo
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
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

	return err
}
