package mongodb

import "github.com/ukane-philemon/bob/db"

// UserNameExists checks if a username exists in the database. Implements
// db.DataStore.
func (m *MongoDB) UserNameExists(username string) (bool, error) {
	return false, nil
}

// EmailExits checks if an email exists in the database. Implements
// db.DataStore.
func (m *MongoDB) EmailExits(email string) (bool, error) {
	return false, nil
}

// CreateUser adds a new user to the database. The username must be unique and
// email must be unique. The password is hashed before being stored. Implements
// db.DataStore.
func (m *MongoDB) CreateUser(username, email string, password []byte) (*db.User, error) {
	return nil, nil
}

// LoginUser logs a user in and returns a nil error if the user exists and the
// password is correct. Implements db.DataStore.
func (m *MongoDB) LoginUser(email string, password []byte) (*db.User, error) {
	return nil, nil
}
