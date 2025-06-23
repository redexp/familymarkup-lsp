package providers

import (
	"go.uber.org/multierr"
	"io"
)

type ReadWriteCloser struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func (r *ReadWriteCloser) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *ReadWriteCloser) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *ReadWriteCloser) Close() error {
	return multierr.Append(r.reader.Close(), r.writer.Close())
}
