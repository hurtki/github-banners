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
	"time"

	"github.com/hurtki/github-banners/api/internal/logger"
)

// Renderer is an infrastrcture layer port, to request renderer service using http
type Renderer struct {
	client          *http.Client
	logger          logger.Logger
	previewEndpoint string
}

// NewRenderer initializes new Renderer that will use given httpClient, logger, and previewEndpoint
// previewEndpoint is an http/s endpoint (example https://renderer/preview/)
func NewRenderer(httpClient *http.Client, logger logger.Logger, previewEndpoint string) *Renderer {
	return &Renderer{
		client:          httpClient,
		logger:          logger.With("service", "renderer-infra"),
		previewEndpoint: previewEndpoint,
	}
}

const (
	requestTimeout = 2 * time.Second
)

// RenderPreview requests renderer service for preview for given bannerInfo
func (c *Renderer) RenderPreview(ctx context.Context, bannerInfo GithubUserBannerInfo) (*GithubBanner, error) {
	fn := "internal.infrastructure.renderer.Renderer.RenderPreview"
	reqBody, err := json.Marshal(bannerInfo.ToBannerPreviewRequest())
	if err != nil {
		c.logger.Error("unexpected error, when marshaling banner preview request", "source", fn, "err", err)
		return nil, ErrCantRequestRenderer
	}

	timeoutContext, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutContext, "POST", c.previewEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		c.logger.Error("unexpected error when preparing request", "source", fn, "err", err)
		return nil, ErrCantRequestRenderer
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)

	if err != nil {
		var urlErr *url.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			// if our timeout context exceeded, so renderer service is unavalible
			return nil, ErrCantRequestRenderer
		case errors.Is(err, context.Canceled):
			return nil, ctx.Err()
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
