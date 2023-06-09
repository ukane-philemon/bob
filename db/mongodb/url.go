package mongodb

import "github.com/ukane-philemon/bob/db"

// SaveURL adds a new URL to the database and returns the shortened URL.
// Implements db.DataStore.
func (m *MongoDB) SaveURL(email string, url string) (string, error) {
	return "", nil
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL. Implements db.DataStore.
func (m *MongoDB) RetrieveURLInfo(short string) (*db.ShortURLInfo, error) {
	return nil, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
// Implements db.DataStore.
func (m *MongoDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	return nil, nil
}
