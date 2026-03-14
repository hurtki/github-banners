-- +goose Up

-- deleting foreign key constraint from repositories table
-- to escape conflicts, because now we will change users table
alter table repositories
drop constraint fk_repository_owner;

-- FIELD username_normalized for github data users table
delete from users a
using users b
where lower(a.username) = lower(b.username)
and a.ctid > b.ctid;

alter table users
drop constraint users_pkey;

drop index if exists idx_users_username;

alter table users
add column username_normalized text;

update users
set username_normalized = lower(username);

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

update repositories
set owner_username_normalized = lower(owner_username);

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




-- +goose Down
-- drop foreign key from repositories
alter table github_data.repositories
drop constraint fk_respository_owner;

-- revert owner_username_normalized back to owner_username
alter table github_data.repositories
add column owner_username text;

update github_data.repositories
set owner_username = owner_username_normalized;

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
