package migrations

import (
	"context"
	"database/sql"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upUsernameNormalizedGithubData, downUsernameNormalizedGithubData)
}

func upUsernameNormalizedGithubData(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
-- deleting foreign key constraint from repositories table
-- to escape conflicts, because now we will change users table
alter table repositories
drop constraint fk_repository_owner;

alter table users
drop constraint users_pkey;

drop index if exists idx_users_username;

alter table users
add column username_normalized text;
`)
	if err != nil {
		return err
	}

	err = normalizeStringRow(ctx, tx, "users", "username", "username_normalized", domain.NormalizeGithubUsername)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
delete from users a
using users b
where a.username_normalized = b.username_normalized
and a.ctid > b.ctid;

alter table users
alter column username_normalized set not null;

alter table users
add constraint users_pkey primary key (username_normalized);

alter table users
alter column username set not null;

create index idx_users_username_normalized on users(username_normalized);

/*
before:
username text primary key

after:
username_normalized text primary key
username text not null
*/

-- normalize repositories table
alter table repositories
add column owner_username_normalized text;
`)

	err = normalizeStringRow(ctx, tx, "repositories", "owner_username", "owner_username_normalized", domain.NormalizeGithubUsername)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
-- now delete ownwer_username column
-- real username should be stored in users table, repositories table can't contain this
drop index if exists idx_repositories_owner_username;
alter table repositories
drop column owner_username;

alter table repositories
alter column owner_username_normalized set not null;

alter table repositories
add constraint fk_repository_owner
    foreign key (owner_username_normalized)
    references users(username_normalized) on delete cascade;

CREATE INDEX idx_repositories_owner_username ON repositories(owner_username_normalized);

/*
before:
owner_username text not null
fk constraint

after:
owner_username_normalized text not null
fk constraint
*/

-- create schema for better separation of gihub data and our service data
-- cause right now "users"
create schema github_data;

alter table users
set schema github_data;

alter table repositories
set schema github_data;
	`)
	return err
}

func downUsernameNormalizedGithubData(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
-- drop foreign key from repositories
alter table github_data.repositories
drop constraint fk_repository_owner;

-- revert owner_username_normalized back to owner_username
alter table github_data.repositories
add column owner_username text;

update github_data.repositories r
set owner_username = u.username
from github_data.users u
where r.owner_username_normalized = u.username_normalized;

drop index if exists idx_repositories_owner_username;
create index idx_repositories_owner_username on github_data.repositories(owner_username);

alter table github_data.repositories
drop column owner_username_normalized;

-- revert users table primary key
alter table github_data.users
drop constraint users_pkey;

alter table github_data.users
drop column username_normalized;

-- recreate original primary key on username
alter table github_data.users
add constraint users_pkey primary key (username);

alter table github_data.users
alter column username set not null;

-- recreate index if needed
create index idx_users_username on github_data.users(username);

-- restore foreign key on repositories.owner_username
alter table github_data.repositories
add constraint fk_repository_owner
    foreign key (owner_username)
    references github_data.users(username) on delete cascade;

-- move tables back to original schema (public)
alter table github_data.users
set schema public;

alter table github_data.repositories
set schema public;

-- optional: drop schema github_data if empty
-- drop schema github_data;
 `)
	return err
}
