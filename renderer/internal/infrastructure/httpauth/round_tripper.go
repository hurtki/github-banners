package httpauth

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SigningRoundTripper struct {
	base 			http.RoundTripper
	serviceName 	string 
	signer 			Signer
	clock 			func() time.Time
}

type Signer interface {
	Sign(data []byte) string
}

func NewAuthHTTPRoundTripper(serviceName string, signer Signer, clock func() time.Time) *SigningRoundTripper {
	if signer == nil {
		panic("signer interface can't be nil, in NewRendererAuthHTTPRoundTripper")
	}
	if clock == nil {
		panic("clock function can't be nil, in NewRendererAuthHTTPRoundTripper")
	}

	return &SigningRoundTripper{
		base: 			http.DefaultTransport, 
		serviceName: 	serviceName,
		signer: 		signer, 
		clock: 			clock,
	}
} 

func (rt *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ts := rt.clock().Unix()

	canonical := strings.Join([]string{
		req.Method, 
		req.URL.Path, 
		strconv.FormatInt(ts, 10), 
		rt.serviceName,
	}, "\n")

	signature := rt.signer.Sign([]byte(canonical))

	r := req.Clone(req.Context())
	r.Header.Set("X-Signature", signature)
	r.Header.Set("X-Timestamp", strconv.FormatInt(ts, 10))
	r.Header.Set("X-Service", rt.serviceName)

	return rt.base.RoundTrip(r)
}
