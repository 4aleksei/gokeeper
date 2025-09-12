// Package datafile - realization file - stream conveer
package datafile

import (
	"errors"

	"github.com/4aleksei/gokeeper/internal/common/aescoder"
	"github.com/4aleksei/gokeeper/internal/common/streams/sources"
	"github.com/4aleksei/gokeeper/internal/common/streams/sources/readwrite"
	"github.com/4aleksei/gokeeper/internal/common/streams/sources/singlefile"
)

type (
	LongtermfileWrite struct {
		success  bool
		closed   bool
		filename string
		writer   *sources.SourceWriter
	}

	LongtermfileRead struct {
		closed   bool
		filename string
		reader   *sources.SourceReader
	}
)

var (
	ErrFileWriteNotSucc = errors.New("error,file write not success")
)

func NewWrite(filename string, key *aescoder.KeyAES) *LongtermfileWrite {
	lw := &LongtermfileWrite{
		writer: sources.CreateWriter(sources.WithDestinationWriter(singlefile.NewWriter(filename)),
			sources.WithSourceWriter(&readwrite.ByteWriter{}),
		),
		filename: filename,
	}
	return lw
}

func NewRead(filename string, key *aescoder.KeyAES) *LongtermfileRead {
	lr := &LongtermfileRead{
		reader: sources.CreateReader(sources.WithDestinationReader(&readwrite.ByteReader{}),
			sources.WithSourceReader(singlefile.NewReader(filename)),
		),
		filename: filename,
	}
	return lr
}

func (l *LongtermfileWrite) Success() {
	l.success = true
}

func (l *LongtermfileWrite) OpenWriter() error {
	return l.writer.OpenWriter()
}

func (l *LongtermfileWrite) CloseWrite() error {
	if l.closed {
		return nil
	}
	l.closed = true
	err := l.writer.CloseWrite()
	if err != nil {
		// Unlink file
		return err
	}
	if !l.success {
		//	Unlink file
		return ErrFileWriteNotSucc
	}
	return nil
}

func (l *LongtermfileWrite) WriteData(b []byte) (int, error) {
	return l.writer.WriteData(b)
}

func (l *LongtermfileRead) ReadData(b []byte) (int, error) {
	return l.reader.ReadData(b)
}

func (l *LongtermfileRead) OpenReader() error {
	return l.reader.OpenReader()
}

func (l *LongtermfileRead) CloseRead() error {
	if l.closed {
		return nil
	}
	l.closed = true
	err := l.reader.CloseRead()
	if err != nil {
		return err
	}
	return nil
}
