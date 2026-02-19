package handlers

import (
	"encoding/json"
	"time"
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

type BannerUpdateStats struct {
	TotalRepos    int            `json:"totalRepos"`
	OriginalRepos int            `json:"originalRepos"`
	ForkedRepos   int            `json:"forkedRepos"`
	TotalStars    int            `json:"totalStars"`
	TotalForks    int            `json:"totalForks"`
	Languages     map[string]int `json:"languages"`
}
