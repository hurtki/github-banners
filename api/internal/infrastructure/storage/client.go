package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hurtki/github-banners/api/internal/logger"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

func NewStorageHTTPClient(roundTripper http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripper,
	}
}

func NewClient(baseURL string, httpClient *http.Client, logger logger.Logger) *Client {
	if baseURL == "" {
		panic("storage: baseURL cannot be empty")
	}
	if httpClient == nil {
		panic("storage: httpClient cannot be nil")
	}
	if logger == nil {
		panic("storage: logger cannot be nil")
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		logger:     logger.With("service", "storage.infrastructure"),
	}
}

func (c *Client) SaveBanner(ctx context.Context, bannerID string, svg string) (string, error) {
	fn := "internal.infrastructure.storage.client.SaveBanner"
	start := time.Now()

	c.logger.Debug("saving banner to storage",
		"source", fn,
		"banner_id", bannerID,
	)

	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	reqBody := SaveRequest{
		URLPath:      bannerID,
		BannerData:   encoded,
		BannerFormat: "svg",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("failed to marshal storage request",
			"source", fn,
			"banner_id", bannerID,
			"err", err,
		)
		return "", fmt.Errorf("marshal save banner request: %w", err)
	}

	url := c.baseURL + "/banners"
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		c.logger.Error("failed to create storage request",
			"source", fn,
			"banner_id", bannerID,
			"err", err,
		)
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("storage request failed",
			"source", fn,
			"banner_id", bannerID,
			"err", err,
		)
		return "", fmt.Errorf("storage request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("storage returned non-success status",
			"source", fn,
			"banner_id", bannerID,
			"status", resp.StatusCode,
			"duration", duration,
			"body", string(respBody),
		)
		return "", fmt.Errorf("storage returned status %d", resp.StatusCode)
	}

	var SaveResp SaveResponse
	if err := json.NewDecoder(resp.Body).Decode(&SaveResp); err != nil {
		c.logger.Error("failed to decode the storage response",
			"source", fn,
			"banner_id", bannerID,
			"duration", duration,
			"err", err,
		)
		return "", fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("banner saved successfully",
		"source", fn,
		"banner_id", bannerID,
		"url", SaveResp.URL,
		"duration", duration,
	)
	return SaveResp.URL, nil
}
