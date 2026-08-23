package custd

import (
	"errors"
	"io"
	"testing"
)

type failingResponseBody struct {
	readErr  error
	closeErr error
}

func (b failingResponseBody) Read([]byte) (int, error) { return 0, b.readErr }
func (b failingResponseBody) Close() error             { return b.closeErr }

func TestReadResponseBodyReportsReadAndCloseErrors(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")

	_, err := readResponseBody(failingResponseBody{readErr: readErr, closeErr: closeErr})
	if !errors.Is(err, readErr) || !errors.Is(err, closeErr) {
		t.Fatalf("readResponseBody error = %v, want read and close errors", err)
	}
	var _ io.ReadCloser = failingResponseBody{}
}
