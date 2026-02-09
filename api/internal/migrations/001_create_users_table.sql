-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY, 
    name TEXT,
    company TEXT,
    location TEXT,
    bio TEXT,
    public_repos_count INT NOT NULL,
    followers_count INT, 
    following_count INT
    fetched_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_users_username ON users(username);

-- +goose Down
DROP TABLE IF EXISTS users;