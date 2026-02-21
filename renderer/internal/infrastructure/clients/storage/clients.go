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

	"github.com/hurtki/github-banners/renderer/internal/logger"
)


type Client struct {
	baseURL 	string 
	httpClient 	*http.Client
	logger 		logger.Logger
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
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		logger: logger.With("component", "storage-client"),
	}
}

func (c *Client) SaveBanner(ctx context.Context, bannerID string, svg string) (string, error) {
	const fn = "infrastructure.clients.storage.SaveBanner"

	start := time.Now()
	
	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	reqBody := SaveRequest{
		URLPath: bannerID,
		BannerData: encoded,
		BannerFormat: "svg",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("%s: marshal request: %w", fn, err)
	}

	url := c.baseURL + "/banners"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("%s: create request: %w", fn, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: do request: %w", fn, err)
	}
	
	defer resp.Body.Close()

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("storage upload failed","status", resp.StatusCode, "banner_id", bannerID, "body", string(respBody), "duration", duration)
		return "", fmt.Errorf("storage returned status %d", resp.StatusCode)
	}

	var saveResp SaveResponse
	if err := json.NewDecoder(resp.Body).Decode(&saveResp); err != nil {
		return "", fmt.Errorf("%s: decode response: %w", fn, err)
	}

	c.logger.Info("banner stored successfully", "banner_id", bannerID, "url", saveResp.URL, "duration", duration)

	return saveResp.URL, nil
}