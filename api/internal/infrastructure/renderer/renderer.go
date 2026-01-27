package renderer

import (
	"context"
)

type Renderer struct {
	client RendererClient
}

type RendererClient interface {
	RequestPreview(context.Context, GithubUserBannerInfo) (*GithubBanner, error)
}

func (a *Renderer) RenderPreview(ctx context.Context, bannerInfo GithubUserBannerInfo) (GithubBanner, error) {
	return GithubBanner{}, nil
}
