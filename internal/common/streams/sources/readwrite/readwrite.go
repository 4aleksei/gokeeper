// Package readwrite - bytes
package readwrite

import (
	"io"
)

type (
	ByteWriter struct {
		w io.Writer
	}

	ByteReader struct {
		r io.Reader
	}
)

func (b *ByteWriter) OpenWriter(w io.Writer) error {
	b.w = w
	return nil
}

func (b *ByteWriter) WriteData(p []byte) (int, error) {
	return b.w.Write(p)
}

func (b *ByteWriter) CloseWrite() error {
	return nil
}

func (b *ByteReader) OpenReader(r io.Reader) error {
	b.r = r
	return nil
}

func (b *ByteReader) ReadData(p []byte) (int, error) {
	return b.r.Read(p)
}

func (b *ByteReader) CloseRead() error {
	return nil
}
