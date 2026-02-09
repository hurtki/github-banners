-- +goose Up
CREATE TABLE IF NOT EXISTS repositories (
    github_id BIGINT PRIMARY KEY, 
    owner_username TEXT NOT NULL, 
    pushed_at TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP NOT NULL,
    language TEXT, 
    stars_count INT NOT NULL,
    is_fork BOOLEAN NOT NULL,
    CONSTRAINT fk_repository_owner
        FOREIGN KEY (owner_username) 
        REFERENCES users(username) ON DELETE CASCADE
);

CREATE INDEX idx_repositories_owner_username ON repositories(owner_username);

-- +goose Down
DROP TABLE IF EXISTS repositories;