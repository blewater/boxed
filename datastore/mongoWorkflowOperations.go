package datastore

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tradeline-tech/workflow/tasks"
)

// Connect the database client
func connect(dbURL, database string, timeout time.Duration) (*mongo.Database, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultTimeoutInSeconds)
	defer cancel()

	// Use for explicit DefaultTimeoutInSeconds for connecting to Mongo
	// Rather than waiting for the Mongo driver default timeout
	clientOptions := options.Client().
		ApplyURI(dbURL).
		SetConnectTimeout(time.Second*DefaultTimeoutInSeconds)

	dbClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err = dbClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return dbClient.Database(database), nil
}

// create a workflow document
func create(ctx context.Context, workflow *tasks.WorkflowType) error {
	if workflow.ID == primitive.NilObjectID {
		workflow.ID = primitive.NewObjectID()
	}

	workflow.CreatedAt = time.Now()

	_, err := dbCollection.InsertOne(ctx, workflow)

	return err
}

// Save a document in databases returns error if document does not exist in the DB
func Save(ctx context.Context, workflow *tasks.WorkflowType) error {
	if db == nil {
		return nil
	}
	if db.Name() == "" {
		return fmt.Errorf("database error")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	workflow.UpdatedAt = time.Now()

	// Create first if not existing
	if workflow.ID == primitive.NilObjectID {
		return create(ctx, workflow)
	}

	filter := bson.M{"_id": workflow.ID}

	opts := options.Replace().SetUpsert(false)
	_, err := dbCollection.ReplaceOne(ctx, filter, workflow, opts)

	return err
}