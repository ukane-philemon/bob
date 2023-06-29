package db

import (
	"crypto/md5"
	"encoding/hex"
)

// GenerateShortURL generates a short URL from a given URL.
func GenerateShortURL(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:URLLength])[:URLLength]
}
