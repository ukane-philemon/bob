package mongodb

import (
	"context"
	"fmt"

	"github.com/ukane-philemon/bob/db"
	"github.com/ukane-philemon/bob/webserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// urlsCollectionName is the name of the collection that stores shortened
	// URLs.
	urlsCollectionName = "urls"
	// urlClicksCollection is the name of the collection that stores shortened
	// URL clicks.
	urlClicksCollection = "url_clicks"
	// usersCollectionName is the name of the collection that stores user
	// information.
	usersCollectionName = "users"
)

const (
	// shortURLKey is the key for the short URL in the database. See:
	// db.ShortURLInfo.ShortURL.
	shortURLKey = "short_url"
	// ownerIDKey is the key for the owner ID in the database. See:
	// db.ShortURLInfo.OwnerID.
	ownerIDKey = "owner_id"
	// emailKey is the key for the user email in the database. See:
	// db.UserInfo.Email.
	emailKey = "email"
	// originalURLKey is the key for the original URL in the database. See:
	// db.ShortURLInfo.OriginalURL.
	originalURLKey = "original_url"
	// usernameKey is the key for the username in the database. See:
	// db.UserInfo.Username.
	usernameKey = "username"
)

type Config struct {
	// DBName is the name of the database.
	DBName string `long:"dbname" env:"MONGODB_DB_NAME" default:"bob" description:"MongoDB database name"`
	// ConnectionUrl is the URL used to connect to the database.
	ConnectionURL string `long:"connectionurl" env:"MONGODB_CONNECTION_URL" description:"MongoDB connection URL"`
}

// MongoDB is the database handler for MongoDB. Implements db.DataStore.
type MongoDB struct {
	ctx context.Context
	db  *mongo.Database
}

// MongoDB implements the db.DataStore interface.
var _ db.DataStore = (*MongoDB)(nil)

// Connect connects to the database and returns a new *MongoDB instance.
func Connect(ctx context.Context, cfg Config) (*MongoDB, error) {
	if cfg.ConnectionURL == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("missing required configuration for MongoDB")
	}

	opts := options.Client().ApplyURI(cfg.ConnectionURL).SetAppName(webserver.AppName)
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection options: %w", err)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := client.Database(cfg.DBName)

	// Create indexes.
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: urlMapKey(shortURLKey), Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err = db.Collection(urlsCollectionName).Indexes().CreateOne(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to create index for urls collection: %w", err)
	}

	model = mongo.IndexModel{
		Keys:    bson.D{{Key: userMapKey(usernameKey), Value: 1}, {Key: userMapKey(emailKey), Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err = db.Collection(usersCollectionName).Indexes().CreateOne(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to create index for users collection: %w", err)
	}

	mdb := &MongoDB{
		ctx: ctx,
		db:  db,
	}

	return mdb, nil
}

// Close ends the connection to the database. Implements db.DataStore.
func (m *MongoDB) Close() error {
	return m.db.Client().Disconnect(m.ctx)
}
