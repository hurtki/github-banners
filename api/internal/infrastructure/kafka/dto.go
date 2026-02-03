package kafka

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type GithubBannerInfoEvent struct {
	EventType	string	
	EventVersion	int 
	ProducedAt	time.Time
	Payload		Payload	
}

type Payload struct {
	Username	string
	BannerType	string
	StoragePath string
	Stats		domain.GithubUserStats
	FetchedAt	time.Time
}
