// Package aescoder
package aescoder

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"io"
)

type (
	KeyAES struct {
		key       []byte
		cipherKey string
	}

	AesReader struct {
		r      io.ReadCloser
		nonce  []byte
		aesgcm cipher.AEAD
	}

	AesWriter struct {
		w      io.Writer
		nonce  []byte
		aesgcm cipher.AEAD
	}
)

func NewAES(pub *rsa.PublicKey) (*KeyAES, error) {
	key, err := generateRandom(2 * aes.BlockSize)
	if err != nil {
		return nil, err
	}
	cipherKeyLoaded, err := rsa.EncryptPKCS1v15(rand.Reader, pub, key)
	if err != nil {
		return nil, err
	}
	return &KeyAES{
		key:       key,
		cipherKey: hex.EncodeToString(cipherKeyLoaded),
	}, nil
}

func (k *KeyAES) GetKey() string {
	return k.cipherKey
}

func NewReader(r io.ReadCloser, key []byte) (*AesReader, error) {
	aesblock, err := aes.NewCipher(key)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	return &AesReader{
		r:      r,
		nonce:  key[len(key)-aesgcm.NonceSize():],
		aesgcm: aesgcm,
	}, nil
}

func (h *AesReader) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	if n > 0 {
		_, err := h.aesgcm.Open(p[0:], h.nonce, p, nil)
		return len(p), err
	}
	return n, err
}

func (h *AesReader) Close() error {
	return nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func NewWriter(w io.Writer, key *KeyAES) (*AesWriter, error) {
	aesblock, err := aes.NewCipher(key.key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	return &AesWriter{
		w:      w,
		aesgcm: aesgcm,
		nonce:  key.key[len(key.key)-aesgcm.NonceSize():],
	}, nil
}

func (a *AesWriter) Write(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	_ = a.aesgcm.Seal(p[0:], a.nonce, p, nil)
	return a.w.Write(p)
}
