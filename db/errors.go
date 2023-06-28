package db

import "errors"

var (
	// ErrorBadRequest is returned for user-facing errors.
	ErrorBadRequest = errors.New("bad request")
)
