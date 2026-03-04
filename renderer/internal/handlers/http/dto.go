package http_handlers

import (
	"time"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/domain/render"
)

type PreviewRequest struct {
	Username   string       `json:"username"`
	BannerType string       `json:"banner_type"`
	Stats      PreviewStats `json:"stats"`
	FetchedAt  time.Time    `json:"fetched_at"`
}

type PreviewStats struct {
	TotalRepos    int            `json:"totalRepos"`
	OriginalRepos int            `json:"originalRepos"`
	ForkedRepos   int            `json:"forkedRepos"`
	TotalStars    int            `json:"totalStars"`
	TotalForks    int            `json:"totalForks"`
	Languages     map[string]int `json:"languages"`
}

func (req PreviewRequest) ToDomainRenderIn() render.RenderIn {
	return render.RenderIn{
		Username:   req.Username,
		BannerType: req.BannerType,
		Stats: domain.GithubUserStats{
			TotalRepos:    req.Stats.TotalRepos,
			OriginalRepos: req.Stats.OriginalRepos,
			ForkedRepos:   req.Stats.ForkedRepos,
			TotalStars:    req.Stats.TotalStars,
			TotalForks:    req.Stats.TotalForks,
			Languages:     req.Stats.Languages,
			FetchedAt:     req.FetchedAt,
		},
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
}
