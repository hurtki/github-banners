package renderer_http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMAC implementation of Signer interface
type HMACSigner struct {
	secret []byte
}

// NewHMACSigner initializes new signer that uses given secret
// panics if secret is blank
func NewHMACSigner(secret []byte) *HMACSigner {
	if len(secret) == 0 {
		panic("secret length in NewHMACSigner shouldn't be 0")
	}
	return &HMACSigner{secret: []byte(secret)}
}

// Sign signs data and returns HEX-coded signature
func (s *HMACSigner) Sign(data []byte) string {
	// using sha256 as hash function
	mac := hmac.New(sha256.New, s.secret)

	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
