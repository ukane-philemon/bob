package webserver

import (
	"crypto/rand"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/ukane-philemon/bob/db"
)

// These are the HTTP codes returned by the API.
const (
	codeOk           = http.StatusOK
	codeBadRequest   = http.StatusBadRequest
	codeInternal     = http.StatusInternalServerError
	codeUnauthorized = http.StatusUnauthorized
	codeFound        = http.StatusFound
)

const (
	// ctxID is the key used to retrieve a user's ID as set by the auth handler
	ctxID = "id"
)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(strings.SplitAfter(email, "@")[1], ".")
}

// randomBytes generates and returns a random byte slice of the specified
// length.
func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// translateDBError translates a database error into an API error.
func translateDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, db.ErrorBadRequest) {
		return errBadRequest(err.Error())
	}

	return errInternal(err)
}
