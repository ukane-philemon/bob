package mongodb

import (
	"fmt"
	"time"

	"github.com/ukane-philemon/bob/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// shortURLKey is the key for the short URL in the database. See:
	// db.ShortURLInfo.ShortURL.
	shortURLKey = "short_url"
	// ownerIDKey is the key for the owner ID in the database. See:
	// db.ShortURLInfo.OwnerID.
	ownerIDKey = "owner_id"
)

// SaveUserURL adds a new URL to the database and returns the shortened URL.
// Implements db.DataStore.
func (m *MongoDB) SaveUserURL(email string, longURL string) (*db.ShortURLInfo, error) {
	return m.createNewShortURL(email, longURL, true)
}

// SaveGuestURL is like SaveUserURL but only for users without an account.
func (m *MongoDB) SaveGuestURL(id string, longURL string) (*db.ShortURLInfo, error) {
	return m.createNewShortURL(id, longURL, true)
}

// createNewShortURL is a shared function between SaveUserURL and SaveGuestURL
// that creates a new short URL. "id" is the user's ID(email) if they are logged in,
// otherwise it is the identifier for the guest user.
func (m *MongoDB) createNewShortURL(id string, longURL string, isGuest bool) (*db.ShortURLInfo, error) {
	if id == "" || longURL == "" {
		return nil, fmt.Errorf("%w: id and url are required", db.ErrorBadRequest)
	}

	if !isGuest && !isValidEmail(id) {
		return nil, fmt.Errorf("%w: invalid email", db.ErrorBadRequest)
	}

	if isGuest {
		// Check if they have reached the maximum number of URLs.
		count, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{ownerIDKey: id})
		if err != nil {
			return nil, fmt.Errorf("error counting documents: %v", err)
		}

		if count >= db.MaxGuestURLs {
			return nil, fmt.Errorf("%w: maximum number of URLs reached", db.ErrorBadRequest)
		}
	} else if res := m.usersCollection().FindOne(m.ctx, bson.M{"email": id}); res.Err() != nil { // Check if user exists.
		return nil, handleUserError(res.Err())
	}

	// Create the short URL.
	urlInfo := &urlInfo{
		ShortURLInfo: &db.ShortURLInfo{
			OwnerID:     ownerIDKey,
			OriginalURL: longURL,
			CreatedAt:   time.Now().UTC().String(),
		},
		IsGuest: isGuest,
	}

	maxTries := 5
	url := longURL
	for maxTries > 0 {
		urlInfo.ShortURL = db.GenerateShortURL(url)
		// Insert the short URL into the database.
		res, err := m.urlsCollection().InsertOne(m.ctx, urlInfo)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("error saving guest URL: %v", err)
		}

		if res != nil && res.InsertedID != nil {
			break
		}

		randomStr, err := randomString(db.URLLength)
		if err != nil {
			return nil, fmt.Errorf("error generating random string: %v", err)
		}

		url = longURL + randomStr
		maxTries--
	}

	return urlInfo.ShortURLInfo, nil

}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL. Implements db.DataStore.
func (m *MongoDB) RetrieveURLInfo(short string) (*db.ShortURLInfo, error) {
	if short == "" {
		return nil, fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	var url *db.ShortURLInfo
	if err := m.urlsCollection().FindOne(m.ctx, bson.M{shortURLKey: short}).Decode(&url); err != nil {
		return nil, handleURLError(err)
	}

	return url, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
// Implements db.DataStore.
func (m *MongoDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	var urls []*db.ShortURLInfo
	cursor, err := m.urlsCollection().Find(m.ctx, bson.M{ownerIDKey: email})
	if err != nil {
		return nil, fmt.Errorf("error retrieving user URLs: %v", err)
	}

	for cursor.Next(m.ctx) {
		var url *db.ShortURLInfo
		if err := cursor.Decode(&url); err != nil {
			return nil, fmt.Errorf("error decoding user URL: %v", err)
		}

		urls = append(urls, url)
	}

	return urls, nil
}

// UpdateShortURL updates the number of clicks for the specified short URL.
func (m *MongoDB) UpdateShortURL(short string) error {
	if short == "" {
		return fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	// Update the short URL clicks in the database.
	filter := bson.M{shortURLKey: short}
	update := bson.M{"$inc": bson.M{"clicks": 1}}
	res, err := m.urlsCollection().UpdateOne(m.ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating short URL: %v", err)
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("%w: short URL does not exist", db.ErrorBadRequest)
	}

	return nil
}

// urlsCollection returns the collection for the short URLs.
func (m *MongoDB) urlsCollection() *mongo.Collection {
	return m.db.Collection(urlsCollectionName)
}

// handleURLError handles errors that occur when retrieving URL information.
func handleURLError(err error) error {
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("%w: URL does not exist", db.ErrorBadRequest)
	}

	return fmt.Errorf("error retrieving URL info: %v", err)
}
