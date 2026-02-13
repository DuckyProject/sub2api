package handler

import (
	"errors"
	"fmt"
	"net/http"
)

func extractMaxBytesError(err error) (*http.MaxBytesError, bool) {
	if maxErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return maxErr, true
	}
	return nil, false
}

func formatBodyLimit(limit int64) string {
	const mb = 1024 * 1024
	if limit >= mb {
		return fmt.Sprintf("%dMB", limit/mb)
	}
	return fmt.Sprintf("%dB", limit)
}

func buildBodyTooLargeMessage(limit int64) string {
	return fmt.Sprintf("Request body too large, limit is %s", formatBodyLimit(limit))
}
