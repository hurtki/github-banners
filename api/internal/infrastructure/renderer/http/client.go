package renderer_http

import (
	"net/http"
)

func NewRendererHTTPClient(roundTripper http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripper,
	}
}
