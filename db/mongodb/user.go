package mongodb

import (
	"fmt"
	"time"

	"github.com/ukane-philemon/bob/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// UsernameExists checks if a username exists in the database. Implements
// db.DataStore.
func (m *MongoDB) UsernameExists(username string) (bool, error) {
	res := m.usersCollection().FindOne(m.ctx, bson.M{userMapKey(usernameKey): username})
	if res.Err() != nil && res.Err() != mongo.ErrNoDocuments {
		return false, res.Err()
	}

	return res.Err() == nil /* false if we got mongo.ErrNoDocuments error */, nil
}

// CreateUser adds a new user to the database. The username must be unique and
// email must be unique. The password is hashed before being stored. Implements
// db.DataStore.
func (m *MongoDB) CreateUser(username, email string, password []byte) error {
	if username == "" || email == "" || password == nil {
		return fmt.Errorf("%w: username, email, and password are required", db.ErrorBadRequest)
	}

	if !db.IsValidEmail(email) {
		return fmt.Errorf("%w: invalid email", db.ErrorBadRequest)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	userInfo := &completeUserInfo{
		UserInfo: &db.UserInfo{
			Username:  username,
			Email:     email,
			Timestamp: time.Now().Unix(),
		},
		Password: hashedPassword,
	}
	if _, err := m.usersCollection().InsertOne(m.ctx, userInfo); err != nil {
		// Check if this is a duplicate error.
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: username or email already exists", db.ErrorBadRequest)
		}

		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

// RetrieveUserInfo fetches information about a user using the email. Implements
// db.DataStore.
func (m *MongoDB) RetrieveUserInfo(email string) (*db.UserInfo, error) {
	if !db.IsValidEmail(email) {
		return nil, fmt.Errorf("%w: a valid email is required", db.ErrorBadRequest)
	}

	res := m.usersCollection().FindOne(m.ctx, bson.M{userMapKey(emailKey): email})
	if res.Err() != nil {
		return nil, handleUserError(res.Err())
	}

	var userInfo *completeUserInfo
	if err := res.Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("error decoding user info: %w", err)
	}

	// Set user's total links
	nLinks, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{urlMapKey(ownerIDKey): email})
	if err != nil {
		return nil, handleURLError(err)
	}
	userInfo.TotalLinks = int(nLinks)

	return userInfo.UserInfo, nil
}

// LoginUser logs a user in and returns a nil error if the user exists and the
// password is correct. Implements db.DataStore.
func (m *MongoDB) LoginUser(email string, password []byte) (*db.UserInfo, error) {
	if !db.IsValidEmail(email) {
		return nil, fmt.Errorf("%w: a valid email is required", db.ErrorBadRequest)
	}

	if len(password) == 0 {
		return nil, fmt.Errorf("%w: password is required", db.ErrorBadRequest)
	}

	res := m.usersCollection().FindOne(m.ctx, bson.M{userMapKey(emailKey): email})
	if res.Err() != nil {
		return nil, handleUserError(res.Err())
	}

	var dbUserInfo *completeUserInfo
	if err := res.Decode(&dbUserInfo); err != nil {
		return nil, fmt.Errorf("error decoding user info: %w", err)
	}

	if bcrypt.CompareHashAndPassword(dbUserInfo.Password, password) != nil {
		return nil, fmt.Errorf("%w: incorrect password", db.ErrorBadRequest)
	}

	// Set user's total links
	nLinks, err := m.urlsCollection().CountDocuments(m.ctx, bson.M{urlMapKey(ownerIDKey): email})
	if err != nil {
		return nil, handleURLError(err)
	}
	dbUserInfo.TotalLinks = int(nLinks)

	return dbUserInfo.UserInfo, nil
}

// usersCollection returns the users collection.
func (m *MongoDB) usersCollection() *mongo.Collection {
	return m.db.Collection(usersCollectionName)
}

// handleUserError handles errors that occur when retrieving a user.
func handleUserError(err error) error {
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("%w: user does not exist", db.ErrorBadRequest)
	}

	return fmt.Errorf("error retrieving user: %w", err)
}

// userMapKey returns a key for the user map.
func userMapKey(key string) string {
	// userKey is the key for the user in the database. See: userInfo.UserInfo.
	return mapKey("user", key)
}
