package mongodb

import (
	"encoding/hex"
	"math/rand"
	"net/mail"
	"strings"
)

// isValidEmail checks if the given email is valid.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(strings.SplitAfter(email, "@")[1], ".")
}

// randomString generates and returns a random string of x2 the specified
// length.
func randomString(len int) (string, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func mapKey(keys ...string) string {
	var mapKey string
	for _, key := range keys {
		if mapKey == "" {
			mapKey = key
		} else {
			mapKey += "." + key
		}
	}
	return mapKey
}
