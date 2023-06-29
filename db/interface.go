package db

const (
	// URLLength is the length of the shortened URL.
	URLLength = 6
	// MaxGuestURLs is the maximum number of URLs a guest can shorten.
	MaxGuestURLs = 2
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
	// CreateNewShortURL adds a new URL to the database and returns the
	// shortened URL. userID will can be any unique identifier for a guest user
	// but it is an email for non-guest users.
	CreateNewShortURL(userID, longURL, customShortURL string, isGuest bool) (*ShortURLInfo, error)
	// UpdateShortURL updates the information for the specified short URL. This
	// method is used for click update and link editing.
	UpdateShortURL(shortURL string, newLongURL string, click *ShortURLClick) error
	// RetrieveURLInfo fetches information about a short URL using the shortened
	// URL.
	RetrieveURLInfo(short string) (*ShortURLInfo, error)
	// RetrieveUserURLs fetches all the shorted URLs for the specified user.
	RetrieveUserURLs(email string) ([]*ShortURLInfo, error)
	// RetrieveShortURLClicks returns a list of complete click information for a
	// short URL.
	RetrieveShortURLClicks(shortURL string) ([]*ShortURLClick, error)
	// ToggleShortLinkStatus enables/disables a short link.
	ToggleShortLinkStatus(shortURL string, disable bool) error
	// Close ends the connection to the database.
	Close() error
}

// User represents a user in the database.
type User struct {
	Username   string `json:"username" bson:"username"`
	Email      string `json:"email" bson:"email"`
	TotalLinks int    `json:"totalLinks" bson:"total_links"`
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
}

// ShortURLInfo represents a short URL in the database.
type ShortURLInfo struct {
	OwnerID     string `json:"ownerID" bson:"owner_id"`
	ShortURL    string `json:"shortUrl" bson:"short_url"`
	OriginalURL string `json:"originalUrl" bson:"original_url"`
	Clicks      int32  `json:"clicks" bson:"clicks"`
	Disabled    bool   `json:"disabled" bson:"disabled"`
	Timestamp   int64  `json:"timestamp" bson:"timestamp"`
}

// ShortURLClick is information about a click on a short URL.
type ShortURLClick struct {
	IP         string `json:"ip" bson:"ip"`
	Browser    string `json:"browser" bson:"browser"`
	Device     string `json:"device" bson:"device"`
	DeviceType string `json:"deviceType" bson:"device_type"`
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
}
