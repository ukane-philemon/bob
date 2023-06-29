package webserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ukane-philemon/bob/db"
	bobdb "github.com/ukane-philemon/bob/db"
)

var (
	tCtx context.Context

	dummyError = errors.New("dummy error")
)

type tServer struct {
	*WebServer
}

// newTServer creates and starts a new server instance. Callers should
// *WebServer.Stop to shutdown the server.
func newTServer(t *testing.T) *tServer {
	var port int
	for port == 0 {
		port = rand.Intn(65535)
	}

	db := newTDB()
	cfg := Config{
		Host: "127.0.0.1",
		Port: fmt.Sprintf("%d", port),
	}

	// Create a new server.
	s, err := New(tCtx, cfg, db)
	if err != nil {
		t.Fatalf("Error creating server: %v", err)
	}

	// Start the server.
	go s.Start()

	return &tServer{s}
}

// sendRequest mimics an actual http request to the server and unmarshals the
// request result into resp. resp must be a pointer to a struct type.
func (ts *tServer) sendRequest(method string, endpoint string, reqBody interface{}, resp interface{}, headers map[string]string) error {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	req := a.Request()
	req.SetRequestURI(fmt.Sprintf("http://%s/%s", ts.addr, endpoint))
	req.Header.SetMethod(method)
	if reqBody != nil {
		req.Header.SetContentType("application/json")
		body, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("json.Marshal error: %w", err)
		}

		req.SetBody(body)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err := a.Parse(); err != nil {
		return err
	}

	_, bodyBytes, errs := a.Bytes()
	var errStr string
	for _, err := range errs {
		errStr += err.Error() + ":"
	}

	if errStr != "" {
		return errors.New(strings.TrimSuffix(errStr, ":"))
	}

	return json.Unmarshal(bodyBytes, resp)
}

func TestMain(m *testing.M) {
	var shutdown context.CancelFunc
	tCtx, shutdown = context.WithCancel(context.Background())
	defer shutdown()
	os.Exit(m.Run())
}

// Mock database for tests.
type tDB struct {
	usernameExists bool
	err            error
	dummyUser      *bobdb.User
	urls           map[string]*bobdb.ShortURLInfo
}

func newTDB() *tDB {
	return &tDB{
		dummyUser: &bobdb.User{
			Username:   "fibrealz",
			Email:      "test@email.com",
			TotalLinks: 20,
			Timestamp:  time.Now().Unix(),
		},
		urls: make(map[string]*bobdb.ShortURLInfo),
	}
}

// UsernameExists checks if a username exists in the database.
func (db *tDB) UsernameExists(username string) (bool, error) {
	return db.usernameExists, db.err
}

// CreateUser adds a new user to the database. The username must be unique
// and email must be unique. The password is hashed before being stored.
func (db *tDB) CreateUser(username, email string, password []byte) error {
	if db.err != nil {
		return db.err
	}

	db.usernameExists = true
	return nil
}

// RetrieveUserInfo fetches information about a user using the email.
func (db *tDB) RetrieveUserInfo(email string) (*bobdb.User, error) {
	if db.err != nil {
		return nil, db.err
	}

	db.dummyUser.Email = email
	return db.dummyUser, nil
}

// LoginUser logs a user in and returns a nil error if the user exists and the
// password is correct.
func (db *tDB) LoginUser(email string, password []byte) (*bobdb.User, error) {
	if db.err != nil {
		return nil, db.err
	}

	if db.dummyUser.Email != email {
		return nil, fmt.Errorf("%w: user does not exist", bobdb.ErrorBadRequest)
	}

	if !bytes.Equal(passwordBytes(dummyUserPassword), password) {
		return nil, fmt.Errorf("%w: incorrect password", bobdb.ErrorBadRequest)
	}

	return db.dummyUser, nil
}

// CreateNewShortURL creates a new short URL. "userID" is the user's email if
// they are logged in, otherwise it is the unique identifier for the guest user.
func (db *tDB) CreateNewShortURL(userID, longURL, customShortURL string, isGuest bool) (*db.ShortURLInfo, error) {
	return nil, errors.New("Not implemented")
}

// UpdateShortURL updates the number of clicks for the specified short URL.
func (db *tDB) UpdateShortURL(shortURL, newLongURL string, click *bobdb.ShortURLClick) error {
	return errors.New("Not implemented")
}

// RetrieveURLInfo fetches information about a short URL using the shortened
// URL.
func (db *tDB) RetrieveURLInfo(short string) (*bobdb.ShortURLInfo, error) {
	return nil, errors.New("Not implemented")
}

// RetrieveUserURLs fetches all the shorted URLs for the specified user.
func (db *tDB) RetrieveUserURLs(email string) ([]*bobdb.ShortURLInfo, error) {
	return nil, errors.New("Not implemented")
}

// RetrieveShortURLClicks returns a list of complete click information for a
// short URL.
func (db *tDB) RetrieveShortURLClicks(shortURL string) ([]*db.ShortURLClick, error) {
	return nil, errors.New("Not implemented")
}

// ToggleShortLinkStatus enables/disables a short link.
func (db *tDB) ToggleShortLinkStatus(shortURL string, disable bool) error {
	return errors.New("Not Implemented")
}

// Close ends the connection to the database.
func (db *tDB) Close() error {
	return nil
}
