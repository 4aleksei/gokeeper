// Package sources - absraction for stream to file conveer
package sources

import (
	"io"
)

type (
	sourceReaderI interface {
		OpenReader() (io.Reader, error)
		CloseRead() error
	}

	middleReaderI interface {
		OpenReader(io.Reader) (io.Reader, error)
		CloseRead() error
	}

	destinationReaderI interface {
		OpenReader(io.Reader) error
		ReadData([]byte) (int, error)
		CloseRead() error
	}

	destinationWriterI interface {
		OpenWriter() (io.Writer, error)
		CloseWrite() error
	}

	middleWriterI interface {
		OpenWriter(io.Writer) (io.Writer, error)
		CloseWrite() error
	}

	sourceWriterI interface {
		OpenWriter(io.Writer) error
		WriteData([]byte) (int, error)
		CloseWrite() error
	}

	SourceReader struct {
		r sourceReaderI
		m []middleReaderI
		f destinationReaderI
	}

	SourceWriter struct {
		w destinationWriterI
		m []middleWriterI
		f sourceWriterI
	}

	OptionSourceWriter interface {
		apply(*SourceWriter)
	}

	destinationWriterIOption struct {
		w destinationWriterI
	}
	middleWriterIOption struct {
		m middleWriterI
	}
	sourceWriterIOption struct {
		f sourceWriterI
	}

	OptionSourceReader interface {
		apply(*SourceReader)
	}

	destinationReaderIOption struct {
		f destinationReaderI
	}
	middleReaderIOption struct {
		m middleReaderI
	}
	sourceReaderIOption struct {
		r sourceReaderI
	}
)

func (w destinationWriterIOption) apply(opts *SourceWriter) {
	opts.w = w.w
}

func WithDestinationWriter(w destinationWriterI) OptionSourceWriter {
	return destinationWriterIOption{w: w}
}
func (m middleWriterIOption) apply(opts *SourceWriter) {
	opts.m = append(opts.m, m.m)
}
func WithMiddleWriter(m middleWriterI) OptionSourceWriter {
	return middleWriterIOption{m: m}
}

func (f sourceWriterIOption) apply(opts *SourceWriter) {
	opts.f = f.f
}
func WithSourceWriter(f sourceWriterI) OptionSourceWriter {
	return sourceWriterIOption{f: f}
}

func CreateWriter(opts ...OptionSourceWriter) *SourceWriter {
	ss := &SourceWriter{}
	for _, o := range opts {
		o.apply(ss)
	}
	return ss
}

func (r destinationReaderIOption) apply(opts *SourceReader) {
	opts.f = r.f
}

func WithDestinationReader(f destinationReaderI) OptionSourceReader {
	return destinationReaderIOption{f: f}
}
func (m middleReaderIOption) apply(opts *SourceReader) {
	opts.m = append(opts.m, m.m)
}
func WithMiddleReader(m middleReaderI) OptionSourceReader {
	return middleReaderIOption{m: m}
}

func (r sourceReaderIOption) apply(opts *SourceReader) {
	opts.r = r.r
}
func WithSourceReader(r sourceReaderI) OptionSourceReader {
	return sourceReaderIOption{r: r}
}

func CreateReader(opts ...OptionSourceReader) *SourceReader {
	ss := &SourceReader{}
	for _, o := range opts {
		o.apply(ss)
	}
	return ss
}

func (sr *SourceReader) OpenReader() error {
	readIo, err := sr.r.OpenReader()
	if err != nil {
		return err
	}

	for _, middle := range sr.m {
		readIo, err = middle.OpenReader(readIo)
		if err != nil {
			_ = sr.CloseRead()
			return err
		}
	}
	sr.f.OpenReader(readIo)
	return nil
}

func (sr *SourceReader) CloseRead() error {
	err := sr.r.CloseRead()
	if err != nil {
		return err
	}
	for _, middle := range sr.m {
		err = middle.CloseRead()
		if err != nil {
			return err
		}
	}
	sr.f.CloseRead()
	return err
}

func (sw *SourceWriter) OpenWriter() error {
	writeIo, err := sw.w.OpenWriter()
	if err != nil {
		return err
	}

	for _, middle := range sw.m {
		writeIo, err = middle.OpenWriter(writeIo)
		if err != nil {
			_ = sw.f.CloseWrite()
			return err
		}
	}
	sw.f.OpenWriter(writeIo)
	return nil
}

func (sw *SourceWriter) CloseWrite() error {
	err := sw.w.CloseWrite()
	if err != nil {
		return err
	}
	for _, middle := range sw.m {
		err = middle.CloseWrite()
		if err != nil {
			return err
		}
	}
	sw.f.CloseWrite()
	return err
}

func (sw *SourceWriter) WriteData(b []byte) (int, error) {
	return sw.f.WriteData(b)
}

func (sr *SourceReader) ReadData(b []byte) (int, error) {
	return sr.f.ReadData(b)
}
