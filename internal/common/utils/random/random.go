package random

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HashPass(p []byte, k string) []byte {
	h := hmac.New(sha256.New, []byte(k))
	dst := h.Sum(p)
	return dst
}
