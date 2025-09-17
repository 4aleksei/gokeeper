// Package aesstream - aes realization in streams  grcp to file or s3
package aesstream

import (
	"io"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
)

type (
	aesWriter struct {
		aesW *aescoder.AesWriter
		key  *aescoder.KeyAES
	}

	aesReader struct {
		aesR *aescoder.AesReader
		key  *aescoder.KeyAES
	}
)

func NewReader(key *aescoder.KeyAES) *aesReader {
	rAes := &aesReader{
		key: key,
	}
	return rAes
}

func NewWriter(key *aescoder.KeyAES) *aesWriter {
	wAes := &aesWriter{
		key: key,
	}
	return wAes
}

func (a *aesReader) OpenReader(r io.Reader) (io.Reader, error) {
	readCloser := io.NopCloser(r)
	rr, err := aescoder.NewReader(readCloser, a.key)
	if err != nil {
		return nil, err
	}
	a.aesR = rr
	return rr, nil
}

func (a *aesReader) CloseRead() error {
	return a.aesR.Close()
}

func (a *aesWriter) OpenWriter(w io.Writer) (io.Writer, error) {
	ww, err := aescoder.NewWriter(w, a.key)
	if err != nil {
		return nil, err
	}
	a.aesW = ww
	return ww, nil
}

func (a *aesWriter) CloseWrite() error {
	return nil
}
