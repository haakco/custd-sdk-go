package custd

import (
	"errors"
	"io"
)

func readResponseBody(body io.ReadCloser) ([]byte, error) {
	data, readErr := io.ReadAll(body)
	closeErr := body.Close()
	return data, errors.Join(readErr, closeErr)
}
