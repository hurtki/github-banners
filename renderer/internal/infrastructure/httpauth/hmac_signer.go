package httpauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HMACSigner struct {
	secret []byte
}

func NewHMACSigner(secret []byte) *HMACSigner {
	if len(secret) == 0 {
		panic("secret length in NewHMACSigner shouldn't be 0")
	}
	return &HMACSigner{secret: []byte(secret)}
}

func (s *HMACSigner) Sign(data []byte) string {
	mac := hmac.New(sha256.New, s.secret)

	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}