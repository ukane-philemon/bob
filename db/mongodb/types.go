package mongodb

import "github.com/ukane-philemon/bob/db"

// completeUserInfo is a wrapper around db.User that includes the password.
type completeUserInfo struct {
	*db.User `bson:"user"`
	Password []byte `bson:"password"`
}

// urlInfo is a wrapper around db.ShortURLInfo that includes whether the URL is
// owned by a guest.
type urlInfo struct {
	*db.ShortURLInfo `bson:"url"`
	IsGuest          bool `bson:"is_guest"`
}

type urlClick struct {
	ShortURL          string `bson:"short_url"`
	*db.ShortURLClick `bson:"click"`
}
