package common

import (
	"context"
	"io"
	"time"
)

type TimeOutReader struct {
	io.Reader
	timeout time.Duration
}

func TimeOutReaderWrap(r io.Reader, d time.Duration) *TimeOutReader {
	return &TimeOutReader{
		Reader:  r,
		timeout: d,
	}
}
func (r *TimeOutReader) Read(buf []byte) (n int, err error) {
	ctx, cancle := context.WithTimeout(context.TODO(), r.timeout)
	defer cancle()
	ErrC := make(chan error)
	go func() {
		n, err = r.Reader.Read(buf)
		ErrC <- err
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-ErrC:
		return
	}
}
