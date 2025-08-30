// Package cryptocerts
package cryptocerts

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
	"os"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
)

type aesReader struct {
	r  io.ReadCloser
	zr *aescoder.AesReader
}

var (
	ErrNoPublic = errors.New("не удалось декодировать публичный ключ")
	ErrNoRSA    = errors.New("не удалось привести к *rsa.PublicKey")
)

func NewAesReader(r io.ReadCloser, privateKeyLoaded *rsa.PrivateKey, aesSkey string) (*aesReader, error) {
	key, err := hex.DecodeString(aesSkey)
	if err != nil {
		return nil, err
	}

	decryptedKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKeyLoaded, key)
	if err != nil {
		return nil, err
	}

	zr, err := aescoder.NewReader(r, decryptedKey)
	if err != nil {
		return nil, err
	}

	return &aesReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *aesReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *aesReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func LoadKey(name string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKeyPEM, err := os.ReadFile(name)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, nil, err
	}
	privateKeyLoaded, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return privateKeyLoaded, &privateKeyLoaded.PublicKey, nil
}

/*func LoadPublicKey(name string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, ErrNoPublic
	}
	publicKeyLoaded, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPublicKeyLoaded, ok := publicKeyLoaded.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNoRSA
	}
	return rsaPublicKeyLoaded, nil
}*/
