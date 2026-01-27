package renderer

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Signer interface {
	Sign(data []byte) (string, error)
}

// RendererAuthHTTPRoundTrip is a custom implementation of http.RoundTripper
// It builds canonical, signs it, and sets headers, so the other service can identify this
type RendererAuthHTTPRoundTripper struct {
	base        http.RoundTripper
	serviceName string
	signer      Signer
	clock       func() time.Time
}

// NewRendererAuthHTTPRoundTripper creates a new auth http round tripper implementation
// It will use signer in order to sign all request to renderer service
// and use clock as part of paylaod ( use time.Now() )
func NewRendererAuthHTTPRoundTripper(serviceName string, signer Signer, clock func() time.Time) *RendererAuthHTTPRoundTripper {
	return &RendererAuthHTTPRoundTripper{
		base:        http.DefaultTransport,
		serviceName: serviceName,
		signer:      signer,
		clock:       clock,
	}
}

func NewRendererHTTPClient(roundTripper http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripper,
	}
}

func (rt *RendererAuthHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ts := rt.clock().Unix()

	// canonical:
	// method[\n]url_path[/n]timestamp[/n]service_name
	canonical := strings.Join([]string{
		req.Method,
		req.URL.Path,
		strconv.FormatInt(ts, 10),
		rt.serviceName,
	}, "\n")

	signature, err := rt.signer.Sign([]byte(canonical))
	if err != nil {
		return nil, err
	}

	// clone, because by convention, only the goroutine that created the request can change it.
	// after creation is kinda immutable
	r := req.Clone(req.Context())
	r.Header.Set("X-Signature", signature)
	r.Header.Set("X-Timestamp", strconv.FormatInt(ts, 10))
	r.Header.Set("X-Service", rt.serviceName)

	return rt.base.RoundTrip(r)
}
