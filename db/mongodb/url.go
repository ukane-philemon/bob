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

// CreateNewShortURL creates a new short URL. "userID" is the user's email if
// they are logged in, otherwise it is the unique identifier for the guest user.
func (m *MongoDB) CreateNewShortURL(userID, longURL, customShortURL string, isGuest bool) (*db.ShortURLInfo, error) {
	if userID == "" || longURL == "" {
		return nil, fmt.Errorf("%w: id and url are required", db.ErrorBadRequest)
	}

	if !isGuest && !db.IsValidEmail(userID) {
		return nil, fmt.Errorf("%w: invalid email", db.ErrorBadRequest)
	}

	if isGuest {
		// Check if they have reached the maximum number of URLs.
		count, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{urlMapKey(ownerIDKey): userID})
		if err != nil {
			return nil, fmt.Errorf("error counting documents: %v", err)
		}

		if count >= db.MaxGuestURLs {
			return nil, fmt.Errorf("%w: maximum number of URLs reached", db.ErrorBadRequest)
		}
	} else if res := m.usersCollection().FindOne(m.ctx, bson.M{userMapKey(emailKey): userID}); res.Err() != nil { // Check if user exists.
		return nil, handleUserError(res.Err())
	}

	newURLInfo := &urlInfo{
		URL: &db.ShortURLInfo{
			OwnerID:     userID,
			OriginalURL: longURL,
			Timestamp:   time.Now().Unix(),
		},
		IsGuest: isGuest,
	}

	customShortURL = strings.TrimSpace(customShortURL)
	if customShortURL != "" {
		newURLInfo.URL.ShortURL = customShortURL
		_, err := m.urlsCollection().InsertOne(m.ctx, newURLInfo)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, fmt.Errorf("%w: custom short URL is already exists %v", db.ErrorBadRequest, err)
			}
			return nil, fmt.Errorf("error saving guest URL: %v", err)
		}

	} else {
		// Check if long URL already exists for this user.
		var oldURLInfo *urlInfo
		err := m.urlsCollection().FindOne(m.ctx, bson.M{urlMapKey(ownerIDKey): userID, urlMapKey(originalURLKey): longURL}).Decode(&oldURLInfo)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("error retrieving URL info: %w", err)
		}

		if oldURLInfo != nil {
			return oldURLInfo.URL, nil
		}

		// Create the short URL.
		maxTries := 5
		url := longURL
		var savedURL bool
		for maxTries > 0 {
			newURLInfo.URL.ShortURL = db.GenerateShortURL(url)
			// Insert the short URL into the database.
			res, err := m.urlsCollection().InsertOne(m.ctx, newURLInfo)
			if err != nil && !mongo.IsDuplicateKeyError(err) {
				return nil, fmt.Errorf("error saving guest URL: %v", err)
			}

			fmt.Printf("%T, %v %v %v", err, err, newURLInfo.URL.ShortURL, mongo.IsDuplicateKeyError(err))

			if res != nil && res.InsertedID != nil {
				savedURL = true
				break
			}

			randomStr, err := db.RandomString(db.URLLength)
			if err != nil {
				return nil, fmt.Errorf("error generating random string: %v", err)
			}

			url = longURL + randomStr
			maxTries--
		}

		if !savedURL {
			return nil, errors.New("failed to save new URL")
		}
	}

	return m.RetrieveURLInfo(newURLInfo.URL.ShortURL)
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL. Implements db.DataStore.
func (m *MongoDB) RetrieveURLInfo(shortURL string) (*db.ShortURLInfo, error) {
	if shortURL == "" {
		return nil, fmt.Errorf("%w: short URL is empty", db.ErrorBadRequest)
	}

	var urlInfo *urlInfo
	if err := m.urlsCollection().FindOne(m.ctx, bson.M{urlMapKey(shortURLKey): shortURL}).Decode(&urlInfo); err != nil {
		return nil, handleURLError(err)
	}

	return urlInfo.URL, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
// Implements db.DataStore.
func (m *MongoDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	var urls []*db.ShortURLInfo
	cursor, err := m.urlsCollection().Find(m.ctx, bson.M{urlMapKey(ownerIDKey): email})
	if err != nil {
		return nil, fmt.Errorf("error retrieving user URLs: %v", err)
	}

	for cursor.Next(m.ctx) {
		var urlInfo *urlInfo
		if err := cursor.Decode(&urlInfo); err != nil {
			return nil, fmt.Errorf("error decoding user URL: %v", err)
		}

		urls = append(urls, urlInfo.URL)
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
	filter := bson.M{urlMapKey(shortURLKey): shortURL}
	update := make(bson.M)
	if click != nil {
		update["$inc"] = bson.M{urlMapKey("clicks"): 1}
		_, err := m.urlClickCollection().InsertOne(m.ctx, &urlClick{
			ShortURL:      shortURL,
			ShortURLClick: click,
		})
		if err != nil {
			return fmt.Errorf("error inserting new click: %w", err)
		}
	} else if newLongURL != "" {
		update["$set"] = bson.M{urlMapKey(originalURLKey): newLongURL}
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
	count, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{urlMapKey(shortURLKey): shortURL})
	if err != nil {
		return nil, handleURLError(err)
	}

	if count == 0 {
		return nil, fmt.Errorf("%w: short url was not found", db.ErrorBadRequest)
	}

	cur, err := m.urlClickCollection().Find(m.ctx, bson.M{shortURLKey: shortURL})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error retrieving link clicks: %w", err)
		}
	}

	var urlClicks []*db.ShortURLClick
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

	filter := bson.M{urlMapKey(shortURLKey): shortURL}
	update := bson.M{"$set": bson.M{urlMapKey("disabled"): disable}}
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

// urlKey returns the key for the specified URL field.
func urlMapKey(key string) string {
	// This key is and must remain consistent with the bson key used in urlInfo.
	return mapKey("url", key)
}
