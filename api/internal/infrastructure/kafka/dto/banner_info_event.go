package dto

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
)

type GithubBannerInfoEvent struct {
	EventType	string	
	EventVersion	int 
	ProducedAt	time.Time
	Payload		renderer.GithubUserBannerInfo
}
