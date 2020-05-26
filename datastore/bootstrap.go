package datastore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	DefaultCollection       = "Workflow"
	DefaultTimeoutInSeconds = 10
)

var (
	dsURL, dsDatabase string
	db                *mongo.Database
	dbCollection      *mongo.Collection
)

func GetDB() *mongo.Database {
	return db
}

func GetDBCollection() *mongo.Collection {
	return dbCollection
}

func NewDataStore(dbURL, dbDatabase, collectionName string) (*mongo.Database, error) {
	dsURL = dbURL
	dsDatabase = dbDatabase

	var err error
	db, err = connect(dbURL, dsDatabase, DefaultTimeoutInSeconds)
	if err != nil {
		return nil, err
	}

	if collectionName != "" {
		dbCollection = db.Collection(collectionName)
	} else {
		dbCollection = db.Collection(DefaultCollection)
	}

	return db, err
}

// Connect the database client
func connect(dbURL, database string, timeout time.Duration) (*mongo.Database, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultTimeoutInSeconds)
	defer cancel()

	// Use for explicit DefaultTimeoutInSeconds for connecting to Mongo
	// Rather than waiting for the Mongo driver default timeout
	clientOptions := options.Client().
		ApplyURI(dbURL).
		SetConnectTimeout(time.Second * DefaultTimeoutInSeconds)

	dbClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err = dbClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return dbClient.Database(database), nil
}
