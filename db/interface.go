package db

const (
	// URLLength is the length of the shortened URL.
	URLLength = 6
	// MaxGuestURLs is the maximum number of URLs a guest can shorten.
	MaxGuestURLs = 5
)

// DataStore is the interface that wraps the basic database operations.
type DataStore interface {
	// UsernameExists checks if a username exists in the database.
	UsernameExists(username string) (bool, error)
	// CreateUser adds a new user to the database. The username must be unique
	// and email must be unique. The password is hashed before being stored.
	CreateUser(username, email string, password []byte) error
	// RetrieveUserInfo fetches information about a user using the email.
	RetrieveUserInfo(email string) (*User, error)
	// LoginUser logs a user in and returns a nil error if the user exists and the
	// password is correct.
	LoginUser(email string, password []byte) (*User, error)
	// SaveUserURL adds a new URL to the database and returns the shortened URL.
	SaveUserURL(email string, longURL string) (*ShortURLInfo, error)
	// SaveGuestURL is like SaveUserURL but only for users without an account.
	SaveGuestURL(id string, longURL string) (*ShortURLInfo, error)
	// UpdateShortURL updates the number of clicks for the specified short URL.
	UpdateShortURL(short string) error
	// RetrieveURLInfo fetches information about a short URL using the shortened
	// URL.
	RetrieveURLInfo(short string) (*ShortURLInfo, error)
	// RetrieveUserURLs fetches all the shorted URLs for the specified user.
	RetrieveUserURLs(email string) ([]*ShortURLInfo, error)
	// Close ends the connection to the database.
	Close() error
}

// User represents a user in the database.
type User struct {
	Username   string `json:"username" bson:"username"`
	Email      string `json:"email" bson:"email"`
	TotalLinks int    `json:"totalLinks" bson:"total_links"`
	CreatedAt  string `json:"createdAt" bson:"created_at"`
}

// ShortURLInfo represents a short URL in the database.
type ShortURLInfo struct {
	OwnerID     string `json:"ownerID" bson:"owner_id"`
	ShortURL    string `json:"shortUrl" bson:"short_url"`
	OriginalURL string `json:"originalUrl" bson:"original_url"`
	Clicks      int32  `json:"clicks" bson:"clicks"`
	CreatedAt   string `json:"createdAt" bson:"created_at"`
}
