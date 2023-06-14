package mongodb

import (
	"fmt"
	"time"

	"github.com/ukane-philemon/bob/db"
)

// SaveUserURL adds a new URL to the database and returns the shortened URL.
// Implements db.DataStore.
func (m *MongoDB) SaveUserURL(email string, url string) (*db.ShortURLInfo, error) {
	return &db.ShortURLInfo{
		OwnerID:     email,
		ShortURL:    "short",
		OriginalURL: url,
		Clicks:      0,
		CreatedAt:   time.Now().String(),
	}, nil
}

// SaveGuestURL is like SaveUserURL but only for users without an account.
func (m *MongoDB) SaveGuestURL(id string, url string) (*db.ShortURLInfo, error) {
	return &db.ShortURLInfo{
		OwnerID:     id,
		ShortURL:    "short",
		OriginalURL: url,
		Clicks:      0,
		CreatedAt:   time.Now().String(),
	}, nil
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL. Implements db.DataStore.
func (m *MongoDB) RetrieveURLInfo(short string) (*db.ShortURLInfo, error) {
	return &db.ShortURLInfo{
		OwnerID:     "email@mail.com",
		ShortURL:    short,
		OriginalURL: "https://www.google.com",
		Clicks:      0,
		CreatedAt:   time.Now().String(),
	}, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
// Implements db.DataStore.
func (m *MongoDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	return []*db.ShortURLInfo{{
		OwnerID:     email,
		ShortURL:    "short",
		OriginalURL: "https://www.google.com",
		Clicks:      0,
		CreatedAt:   time.Now().String(),
	}}, nil
}

// UpdateShortURL updates the number of clicks for the specified short URL.
func (m *MongoDB) UpdateShortURL(short string) error {
	fmt.Println("UpdateShortURL", short)
	return nil
}
