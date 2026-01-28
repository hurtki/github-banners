package renderer_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
	"github.com/hurtki/github-banners/api/internal/logger"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

type StringBody struct {
	Reader io.Reader
}

func (b *StringBody) Read(data []byte) (int, error) { return b.Reader.Read(data) }
func (b *StringBody) Close() error                  { return nil }

func newHeader(headers map[string]string) http.Header {
	var header http.Header = make(map[string][]string)
	for key, val := range headers {
		header.Add(key, val)
	}
	return header
}

func TestRenderer_RenderPreview(t *testing.T) {
	httpmock.Activate(t)

	apiUlr := "https://renderer/preview"

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		httpClient   *http.Client
		logger       logger.Logger
		endpoint     string
		responseBody string
		httpResponse *http.Response
		// Named input parameters for target function.
		bannerInfo renderer.GithubUserBannerInfo
		want       *renderer.GithubBanner
		wantedErr  error
	}{
		{
			name:       "base",
			httpClient: http.DefaultClient,
			logger:     logger.NewLogger("info", "json"),
			endpoint:   apiUlr,
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Header: newHeader(map[string]string{
					"Content-Type": "image/svg+xml",
				}),
				Body: &StringBody{strings.NewReader("<svg></svg>")},
			},
			bannerInfo: renderer.GithubUserBannerInfo{},
			want:       &renderer.GithubBanner{Username: "", BannerType: "", Banner: []byte("<svg></svg>")},
			wantedErr:  nil,
		},
		{
			name:       "wrong-format",
			httpClient: http.DefaultClient,
			logger:     logger.NewLogger("info", "json"),
			endpoint:   apiUlr,
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Header: newHeader(map[string]string{
					"Content-Type": "image/jpeg",
				}),
				Body: &StringBody{strings.NewReader("some jpeg stuff")},
			},
			bannerInfo: renderer.GithubUserBannerInfo{},
			want:       nil,
			wantedErr:  renderer.ErrCantRequestRenderer,
		},
		{
			name:       "bad-request-status",
			httpClient: http.DefaultClient,
			logger:     logger.NewLogger("info", "json"),
			endpoint:   apiUlr,
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     newHeader(map[string]string{}),
				Body:       &StringBody{strings.NewReader("{\"error\":\"negative repos count\"}")},
			},
			bannerInfo: renderer.GithubUserBannerInfo{},
			want:       nil,
			wantedErr:  renderer.ErrBadPreviewRequest,
		},
	}
	for _, tt := range tests {
		httpmock.Reset()
		httpmock.RegisterResponder("POST", tt.endpoint, httpmock.ResponderFromResponse(tt.httpResponse))

		t.Run(tt.name, func(t *testing.T) {
			c := renderer.NewRenderer(tt.httpClient, tt.logger, tt.endpoint)
			got, gotErr := c.RenderPreview(context.Background(), tt.bannerInfo)
			require.Equal(t, tt.wantedErr, gotErr)
			require.Equal(t, tt.want, got)
		})
	}
}
