-- +goose Up
alter table banners
drop constraint banners_github_username_banner_type_key;

alter table banners
add column github_username_normalized text;

update banners
set github_username_normalized = lower(github_username);

alter table banners
alter column github_username_normalized set not null;

alter table banners
drop column github_username;

-- deduplicate
delete from banners a
using banners b
-- same lowered ( normalized username )
where a.github_username_normalized = b.github_username_normalized
-- same banner type
and a.banner_type = b.banner_type;
-- their ctids are different ( different rows )
and a.ctid > b.ctid

alter table banners
add constraint banners_github_username_normalized_banner_type_key unique (github_username_normalized, banner_type);

create index idx_banners_username_normalized
on banners (github_username_normalized);

-- +goose Down
alter table banners
drop constraint banners_github_username_normalized_banner_type_key;

drop index if exists idx_banners_username_normalized;

alter table banners
add column github_username text;

update banners
set github_username = github_username_normalized;

alter table banners
alter column github_username set not null;

alter table banners
drop column github_username_normalized;

alter table banners
add constraint banners_github_username_banner_type_key
unique (github_username, banner_type);
