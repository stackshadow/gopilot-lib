package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// ComputeHmac256 will return an sha256 HMAC-hash as base64
func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
