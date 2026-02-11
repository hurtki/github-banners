-- +goose Up
CREATE TABLE IF NOT EXISTS banners (
    github_username TEXT PRIMARY KEY, 
    banner_type TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_banners_github_username ON banners(github_username);

-- +goose Down
DROP TABLE IF EXISTS banners;