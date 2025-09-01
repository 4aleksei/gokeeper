// Package cryptocerts
package cryptocerts

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var (
	ErrNoPublic = errors.New("не удалось декодировать публичный ключ")
	ErrNoRSA    = errors.New("не удалось привести к *rsa.PublicKey")
)

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

func GenerateKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
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
