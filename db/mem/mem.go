package mem

import (
	"fmt"
	"sync"
	"time"

	"github.com/ukane-philemon/bob/db"
	"golang.org/x/crypto/bcrypt"
)

// MemDB is an in-memory database.
type MemDB struct {
	mtx        sync.RWMutex
	urls       map[string]*db.ShortURLInfo
	users      map[string]*db.UserInfo
	urlClicks  map[string][]*db.ShortURLClick
	hashedPass map[string][]byte
	err        error
}

// MemDB implements the db.DataStore interface.
var _ db.DataStore = (*MemDB)(nil)

// New returns a new *MemDB instance.
func New() *MemDB {
	return &MemDB{
		urls:       make(map[string]*db.ShortURLInfo),
		users:      make(map[string]*db.UserInfo),
		urlClicks:  make(map[string][]*db.ShortURLClick),
		hashedPass: make(map[string][]byte),
	}
}

// UsernameExists checks if a username exists in the database.
func (m *MemDB) UsernameExists(username string) (bool, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return false, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// CreateUser adds a new user to the database. The username must be unique
// and email must be unique. The password is hashed before being stored.
func (m *MemDB) CreateUser(username, email string, password []byte) error {
	if m.err != nil {
		err := m.err
		m.err = nil
		return err
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	if _, ok := m.users[email]; ok {
		return fmt.Errorf("%w: email already exists", db.ErrorBadRequest)
	}

	m.users[email] = &db.UserInfo{
		Username:  username,
		Email:     email,
		Timestamp: time.Now().Unix(),
	}

	var err error
	m.hashedPass[email], err = bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt.GenerateFromPassword error: %w", err)
	}

	return nil
}

// RetrieveUserInfo fetches information about a user using the email.
func (m *MemDB) RetrieveUserInfo(email string) (*db.UserInfo, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	user, ok := m.users[email]
	if !ok {
		return nil, fmt.Errorf("%w: email does not exist", db.ErrorBadRequest)
	}

	u := *user
	return &u, nil
}

// LoginUser logs a user in and returns a nil error if the user exists and the
// password is correct.
func (m *MemDB) LoginUser(email string, password []byte) (*db.UserInfo, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	user, ok := m.users[email]
	if !ok {
		return nil, fmt.Errorf("%w: user does not exist", db.ErrorBadRequest)
	}

	userPass, ok := m.hashedPass[email]
	if !ok {
		return nil, fmt.Errorf("%w: user does not exist", db.ErrorBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword(userPass, password); err != nil {
		return nil, fmt.Errorf("%w: incorrect password", db.ErrorBadRequest)
	}

	return user, nil
}

// CreateNewShortURL adds a new URL to the database and returns the
// shortened URL. userID will can be any unique identifier for a guest user
// but it is an email for non-guest users.
func (m *MemDB) CreateNewShortURL(userID, longURL, customShortURL string, isGuest bool) (*db.ShortURLInfo, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if !isGuest && !db.IsValidEmail(userID) {
		return nil, fmt.Errorf("%w: invalid email", db.ErrorBadRequest)
	}

	if customShortURL != "" {
		if _, ok := m.urls[customShortURL]; ok {
			return nil, fmt.Errorf("%w: short URL already exists", db.ErrorBadRequest)
		}
	}

	var err error
	shortURL := customShortURL
	if shortURL == "" {
		shortURL, err = db.RandomString(3)
		if err != nil {
			return nil, err
		}
	}

	m.urls[shortURL] = &db.ShortURLInfo{
		OriginalURL: longURL,
		ShortURL:    shortURL,
		OwnerID:     userID,
		Timestamp:   time.Now().Unix(),
	}

	return m.urls[shortURL], nil
}

// UpdateShortURL updates the information for the specified short URL. This
// method is used for click update and link editing.
func (m *MemDB) UpdateShortURL(shortURL string, newLongURL string, click *db.ShortURLClick) error {
	if m.err != nil {
		err := m.err
		m.err = nil
		return err
	}

	if newLongURL == "" && click == nil {
		return fmt.Errorf("%w: nothing to update", db.ErrorBadRequest)
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	url := m.urls[shortURL]
	if url == nil {
		return fmt.Errorf("%w: short URL not found", db.ErrorBadRequest)
	}

	if newLongURL != "" {
		url.OriginalURL = newLongURL
	} else if click != nil {
		m.urlClicks[shortURL] = append(m.urlClicks[shortURL], click)
	}

	return nil
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL.
func (m *MemDB) RetrieveURLInfo(short string) (*db.ShortURLInfo, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()
	url := m.urls[short]
	if url == nil {
		return nil, fmt.Errorf("%w: short URL not found", db.ErrorBadRequest)
	}

	l := *url
	return &l, nil
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
func (m *MemDB) RetrieveUserURLs(email string) ([]*db.ShortURLInfo, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()
	var urls []*db.ShortURLInfo
	for _, url := range m.urls {
		if url.OwnerID == email {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

// RetrieveShortURLClicks returns a list of complete click information for a
// short URL.
func (m *MemDB) RetrieveShortURLClicks(shortURL string) ([]*db.ShortURLClick, error) {
	if m.err != nil {
		err := m.err
		m.err = nil
		return nil, err
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return m.urlClicks[shortURL], nil
}

// ToggleShortLinkStatus enables/disables a short link.
func (m *MemDB) ToggleShortLinkStatus(shortURL string, disable bool) error {
	if m.err != nil {
		err := m.err
		m.err = nil
		return err
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	if url, ok := m.urls[shortURL]; ok {
		url.Disabled = disable
		return nil
	}
	return fmt.Errorf("%w: short URL does not exist", db.ErrorBadRequest)
}

// Close ends the connection to the database.
func (m *MemDB) Close() error {
	// Empty the db to free up memory.
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.urls = make(map[string]*db.ShortURLInfo)
	m.users = make(map[string]*db.UserInfo)
	m.urlClicks = make(map[string][]*db.ShortURLClick)
	m.hashedPass = make(map[string][]byte)
	return nil
}

// SetError is used by tests to simulate errors.
func (m *MemDB) SetError(err error) {
	m.err = err
}
