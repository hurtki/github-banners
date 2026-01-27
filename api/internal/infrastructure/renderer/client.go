package renderer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hurtki/github-banners/api/internal/logger"
)

var (
	ErrCantMarshalRequest  = errors.New("can't marshal request")
	ErrCantRequestRenderer = errors.New("can't request renderer")
	ErrBadPreviewRequest   = errors.New("bad preview request")
)

type RendererClientImpl struct {
	client   *http.Client
	logger   logger.Logger
	endpoint string
}

func NewRendererClient(httpClient *http.Client, logger logger.Logger, endpoint string) *RendererClientImpl {
	return &RendererClientImpl{
		client:   http.DefaultClient,
		logger:   logger,
		endpoint: endpoint,
	}
}

func (c *RendererClientImpl) RequestPreview(ctx context.Context, bannerInfo GithubUserBannerInfo) (*GithubBanner, error) {
	fn := "internal.infrastructure.renderer.RendererClient.RequestPreview"
	reqBody, err := json.Marshal(bannerInfo.ToBannerPreviewRequest)
	if err != nil {
		c.logger.Error("can't marshal banner info to json", "source", fn, "err", err)
		return nil, ErrCantMarshalRequest
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		c.logger.Error("unexpected error when preparing request", "source", fn, "err", err)
		return nil, ErrCantRequestRenderer
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)

	if err != nil {
		var urlErr *url.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			return nil, err
		case errors.As(err, &urlErr):
			c.logger.Error("network error occured, when requesting renderer service", "source", fn, "err", urlErr)
			return nil, ErrCantRequestRenderer
		default:
			c.logger.Error("unexpected error, when requesting renderer service", "source", fn, "err", err)
			return nil, ErrCantRequestRenderer
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch {
		case res.StatusCode == http.StatusUnauthorized:
			c.logger.Error("unauthorized status code returned from renderer service", "source", fn)
			return nil, ErrCantRequestRenderer
		case res.StatusCode >= 400 && res.StatusCode < 500:
			return nil, ErrBadPreviewRequest
		default:
			c.logger.Error("unexpected status code from renderer service", "source", fn, "code", res.StatusCode)
			return nil, ErrCantRequestRenderer
		}
	}
	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/svg+xml") {
		c.logger.Error("unexpected content type from renderer service", "source", fn, "content-type", ct)
		return nil, ErrCantRequestRenderer
	}
	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		c.logger.Error("unexpected error, when reading response body", "source", fn, "err", err)
		return nil, ErrCantRequestRenderer
	}
	return &GithubBanner{
		Username:   bannerInfo.Username,
		BannerType: bannerInfo.BannerType,
		Banner:     resBody,
	}, nil
}
