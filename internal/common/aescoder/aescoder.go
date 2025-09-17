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

func DecodeAESKey(prv *rsa.PrivateKey, hexKey string) (*KeyAES, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	decryptedKey, err := rsa.DecryptPKCS1v15(rand.Reader, prv, key)

	if err != nil {
		return nil, err
	}
	return &KeyAES{
		key:       decryptedKey,
		cipherKey: hexKey,
	}, nil
}

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

func NewReader(r io.ReadCloser, key *KeyAES) (*AesReader, error) {
	aesblock, err := aes.NewCipher(key.key)

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
		nonce:  key.key[len(key.key)-aesgcm.NonceSize():],
		aesgcm: aesgcm,
	}, nil
}

func (h *AesReader) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	if n > 0 {
		np, err := h.aesgcm.Open(p[:0], h.nonce, p[:n], nil)
		return len(np), err
	}
	return n, err

	/*
		n, err := h.r.Read(p)
		fmt.Println("READ FILE", n)
		if err != nil {
			return n, err
		}

		if n > 0 {
			np, err := h.aesgcm.Open(nil, h.nonce, p[:n], nil)
			if err != nil {
				return 0, err
			}
			copy(p, np)
			return len(p), nil
		}
		return n, err*/
}

func (h *AesReader) ReadOne(p []byte) ([]byte, error) {
	n, _ := h.r.Read(p)
	if n > 0 {
		np, err := h.aesgcm.Open(nil, h.nonce, p, nil)
		if err != nil {
			return nil, err
		}
		return np, err
	}
	return p, nil
}

func (h *AesReader) Close() error {
	h.r.Close()
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
	k := a.aesgcm.Seal(nil, a.nonce, p, nil)
	return a.w.Write(k)
}

func (a *AesWriter) WriteOne(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	k := a.aesgcm.Seal( /*p[0:]*/ nil, a.nonce, p, nil)
	return a.w.Write(k)
}
