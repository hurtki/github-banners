package renderer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HMAC implementation of Signer interface
type HMACSigner struct {
	secret []byte
}

func NewHMACSigner(secret string) *HMACSigner {
	return &HMACSigner{secret: []byte(secret)}
}

// Sign signs data and returns HEX-coded signature
func (s *HMACSigner) Sign(data []byte) (string, error) {
	if len(s.secret) == 0 {
		return "", errors.New("empty secret")
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}
