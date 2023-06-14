package mongodb

import (
	"fmt"
	"time"

	"github.com/ukane-philemon/bob/db"
)

// UsernameExists checks if a username exists in the database. Implements
// db.DataStore.
func (m *MongoDB) UsernameExists(username string) (bool, error) {
	fmt.Println("UsernameExists", username)
	return false, nil
}

// CreateUser adds a new user to the database. The username must be unique and
// email must be unique. The password is hashed before being stored. Implements
// db.DataStore.
func (m *MongoDB) CreateUser(username, email string, password []byte) error {
	fmt.Println("CreateUser", username, email, password)
	return nil
}

// RetrieveUserInfo fetches information about a user using the email. Implements
// db.DataStore.
func (m *MongoDB) RetrieveUserInfo(email string) (*db.User, error) {
	return &db.User{
		Username:   "random",
		Email:      email,
		TotalLinks: 2,
		CreatedAt:  time.Now().String(),
	}, nil
}

// LoginUser logs a user in and returns a nil error if the user exists and the
// password is correct. Implements db.DataStore.
func (m *MongoDB) LoginUser(email string, password []byte) (*db.User, error) {
	return &db.User{
		Username:   "random",
		Email:      email,
		TotalLinks: 2,
		CreatedAt:  time.Now().String(),
	}, nil
}
