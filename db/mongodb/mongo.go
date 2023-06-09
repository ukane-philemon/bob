package mongodb

import (
	"context"

	"github.com/ukane-philemon/bob/db"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB is the database handler for MongoDB. Implements db.DataStore.
type MongoDB struct {
	ctx context.Context
	db  *mongo.Database
}

// MongoDB implements the db.DataStore interface.
var _ db.DataStore = (*MongoDB)(nil)

// Connect connects to the database and returns a new *MongoDB instance.
func Connect(ctx context.Context, connectionUrl string) (*MongoDB, error) {
	return nil, nil
}

// Close ends the connection to the database. Implements db.DataStore.
func (m *MongoDB) Close() error {
	return nil
}
