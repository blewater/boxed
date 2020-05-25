package datastore
import (

"go.mongodb.org/mongo-driver/mongo"
)

const (
	DefaultCollection = "Workflow"
	DefaultTimeoutInSeconds = 10
)

var (
	dsURL, dsDatabase string
	db *mongo.Database
	dbCollection *mongo.Collection
)

func NewDataStore(dbURL, dbDatabase, collectionName string) (*mongo.Database, error){
	dsURL = dbURL
	dsDatabase = dbDatabase

	var err error
	db, err = connect(dbURL, dsDatabase, DefaultTimeoutInSeconds)
	if err != nil {
		return nil, err
	}

	if collectionName != "" {
		dbCollection = db.Collection(collectionName)
	}else {
		dbCollection = db.Collection(DefaultCollection)
	}

	return db, err
}



