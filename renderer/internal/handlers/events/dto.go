package events

import (
	"encoding/json"
	"time"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/domain/render"
)

type Message struct {
	Key   []byte
	Value []byte
}

type BannerUpdateEvent struct {
	EventType    string           `json:"event_type"`
	EventVersion int              `json:"event_version"`
	ProducedAt   time.Time        `json:"produced_at"` // RFC3339
	Payload      BannerUpdateInfo `json:"payload"`
}

func (e BannerUpdateEvent) String() string {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "<error marshaling EventDTO>"
	}
	return string(b)
}

type BannerUpdateInfo struct {
	Username    string            `json:"username"`
	BannerType  string            `json:"banner_type"`
	StoragePath string            `json:"storage_path"`
	FetchedAt   time.Time         `json:"fetched_at"` // RFC3339
	Stats       BannerUpdateStats `json:"stats"`
}

func (i BannerUpdateInfo) ToDomainInUpdateBannerIn() render.UpdateBannerIn {
	return render.UpdateBannerIn{
		Username:   i.Username,
		BannerType: i.BannerType,
		URLPath:    i.StoragePath,
		Stats: domain.GithubUserStats{
			TotalRepos:    i.Stats.TotalRepos,
			OriginalRepos: i.Stats.OriginalRepos,
			ForkedRepos:   i.Stats.ForkedRepos,
			TotalStars:    i.Stats.TotalStars,
			TotalForks:    i.Stats.TotalForks,
			Languages:     i.Stats.Languages,
			FetchedAt:     i.FetchedAt,
		},
	}
}

type BannerUpdateStats struct {
	TotalRepos    int            `json:"total_repos"`
	OriginalRepos int            `json:"original_repos"`
	ForkedRepos   int            `json:"forked_repos"`
	TotalStars    int            `json:"total_stars"`
	TotalForks    int            `json:"total_forks"`
	Languages     map[string]int `json:"languages"`
}
