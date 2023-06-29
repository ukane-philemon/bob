package mongodb

import (
	"errors"
	"fmt"
	"strings"
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

// CreateNewShortURL creates a new short URL. "userID" is the user's email if
// they are logged in, otherwise it is the unique identifier for the guest user.
func (m *MongoDB) CreateNewShortURL(userID, longURL, customShortURL string, isGuest bool) (*db.ShortURLInfo, error) {
	if userID == "" || longURL == "" {
		return nil, fmt.Errorf("%w: id and url are required", db.ErrorBadRequest)
	}

	if !isGuest && !isValidEmail(userID) {
		return nil, fmt.Errorf("%w: invalid email", db.ErrorBadRequest)
	}

	if isGuest {
		// Check if they have reached the maximum number of URLs.
		count, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{mapKey("url", ownerIDKey): userID})
		if err != nil {
			return nil, fmt.Errorf("error counting documents: %v", err)
		}

		if count >= db.MaxGuestURLs {
			return nil, fmt.Errorf("%w: maximum number of URLs reached", db.ErrorBadRequest)
		}
	} else if res := m.usersCollection().FindOne(m.ctx, bson.M{"email": userID}); res.Err() != nil { // Check if user exists.
		return nil, handleUserError(res.Err())
	}

	newURLInfo := &urlInfo{
		ShortURLInfo: &db.ShortURLInfo{
			OwnerID:     userID,
			OriginalURL: longURL,
			Timestamp:   time.Now().Unix(),
		},
		IsGuest: isGuest,
	}

	customShortURL = strings.TrimSpace(customShortURL)
	if customShortURL != "" {
		newURLInfo.ShortURL = customShortURL
		nLinks, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{mapKey("url", shortURLKey): customShortURL})
		if err != nil {
			return nil, handleURLError(err)
		}

		if nLinks > 0 {
			return nil, fmt.Errorf("%w: custom short URL is already exists", db.ErrorBadRequest)
		}

		_, err = m.urlsCollection().InsertOne(m.ctx, newURLInfo)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, fmt.Errorf("%w: custom short URL is already exists", db.ErrorBadRequest)
			}

			return nil, fmt.Errorf("error saving guest URL: %v", err)
		}

	} else {
		// Check if long URL already exists for this user.
		var oldURL *urlInfo
		if err := m.urlsCollection().FindOne(m.ctx, bson.M{mapKey("url", ownerIDKey): userID, "original_url": longURL}).Decode(&oldURL); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("error retireving URL info: %w", err)
		}

		if oldURL != nil {
			return oldURL.ShortURLInfo, nil
		}

		// Create the short URL.
		maxTries := 5
		url := longURL
		newURLInfo.ShortURL = db.GenerateShortURL(url)
		for maxTries > 0 {
			// Insert the short URL into the database.
			res, err := m.urlsCollection().InsertOne(m.ctx, newURLInfo)
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
			newURLInfo.ShortURL = db.GenerateShortURL(url)
			maxTries--
		}
	}

	return m.RetrieveURLInfo(newURLInfo.ShortURL)
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL. Implements db.DataStore.
func (m *MongoDB) RetrieveURLInfo(short string) (*db.ShortURLInfo, error) {
	if short == "" {
		return nil, fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	var url *db.ShortURLInfo
	if err := m.urlsCollection().FindOne(m.ctx, bson.M{mapKey("url", shortURLKey): short}).Decode(&url); err != nil {
		return nil, handleURLError(err)
	}

	return url, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
// Implements db.DataStore.
func (m *MongoDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	var urls []*db.ShortURLInfo
	cursor, err := m.urlsCollection().Find(m.ctx, bson.M{mapKey("url", ownerIDKey): email})
	if err != nil {
		return nil, fmt.Errorf("error retrieving user URLs: %v", err)
	}

	for cursor.Next(m.ctx) {
		var url *urlInfo
		if err := cursor.Decode(&url); err != nil {
			return nil, fmt.Errorf("error decoding user URL: %v", err)
		}

		urls = append(urls, url.ShortURLInfo)
	}

	return urls, nil
}

// UpdateShortURL updates the information for the specified short URL. This
// method is used for click update and link editing.
func (m *MongoDB) UpdateShortURL(shortURL, newLongURL string, click *db.ShortURLClick) error {
	if shortURL == "" {
		return fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	// Update the short URL clicks in the database.
	filter := bson.M{mapKey("url", shortURLKey): shortURL}
	update := make(bson.M)
	if click != nil {
		update["$inc"] = bson.M{"clicks": 1}
		_, err := m.urlClickCollection().InsertOne(m.ctx, &urlClick{
			ShortURL:      shortURL,
			ShortURLClick: click,
		})
		if err != nil {
			return fmt.Errorf("error inserting new click: %w", err)
		}
	} else if newLongURL != "" {
		update["$set"] = bson.M{"original_url": newLongURL}
	}

	res, err := m.urlsCollection().UpdateOne(m.ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating short URL: %v", err)
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("%w: short URL does not exist", db.ErrorBadRequest)
	}

	return nil
}

// RetrieveShortURLClicks returns a list of complete click information for a
// short URL.
func (m *MongoDB) RetrieveShortURLClicks(shortURL string) ([]*db.ShortURLClick, error) {
	if shortURL == "" {
		return nil, fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	// Confirm link exists
	count, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{mapKey("url", shortURLKey): shortURL})
	if err != nil {
		return nil, handleURLError(err)
	}

	if count < 1 {
		return nil, fmt.Errorf("%w: short url was not found", db.ErrorBadRequest)
	}

	var urlClicks []*db.ShortURLClick
	cur, err := m.urlClickCollection().Find(m.ctx, bson.M{shortURLKey: shortURL})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error retreiveing link clicks: %w", err)
		}
	}

	for cur.Next(m.ctx) {
		var click *urlClick
		if err = cur.Decode(&click); err != nil {
			return nil, fmt.Errorf("cursor.Decode error: %w", err)
		}
		urlClicks = append(urlClicks, click.ShortURLClick)
	}

	return urlClicks, nil
}

// ToggleShortLinkStatus enables/disables a short link.
func (m *MongoDB) ToggleShortLinkStatus(shortURL string, disable bool) error {
	if shortURL == "" {
		return fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	filter := bson.M{mapKey("url", shortURLKey): shortURL}
	update := bson.M{"$set": bson.M{"disabled": disable}}
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

// urlClickCollection returns the collection for short URL clicks.
func (m *MongoDB) urlClickCollection() *mongo.Collection {
	return m.db.Collection(urlClicksCollection)
}

// handleURLError handles errors that occur when retrieving URL information.
func handleURLError(err error) error {
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("%w: URL does not exist", db.ErrorBadRequest)
	}

	return fmt.Errorf("error retrieving URL info: %v", err)
}
