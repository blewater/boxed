package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tradeline-tech/workflow/datastore"
)

// create a workflow document
func create(ctx context.Context, workflow *Tasks) error {
	if workflow.ID == primitive.NilObjectID {
		workflow.ID = primitive.NewObjectID()
	}

	workflow.CreatedAt = time.Now()

	_, err := datastore.GetDBCollection().InsertOne(ctx, workflow)

	return err
}

// Read returns a document based on the filter
func Read(ctx context.Context, dbConn *mongo.Database, filter bson.M) (*Tasks, error) {
	workflow := &Tasks{}

	res := datastore.GetDBCollection().FindOne(ctx, filter)
	err := res.Decode(workflow)

	return workflow, err
}

// LoadByName returns the database organization workflow database
func LoadByName(ctx context.Context, workflowNameKey string) (*Tasks, error) {
	filter := bson.M{"workflowNameKey": workflowNameKey}

	dbWorkflow, err := Read(ctx, datastore.GetDB(), filter)

	if err != nil && err.Error() != "mongo: no documents in result" {
		return nil, err
	}

	if dbWorkflow == nil {
		return nil, errors.New("failed to load workflow from mongo or id found to be nil")
	}
	if dbWorkflow.ID == primitive.NilObjectID {
		return nil, fmt.Errorf("workflow from mongo with nil id: %v", dbWorkflow.ID.Hex())
	}

	return dbWorkflow, nil
}

// Save a document in databases returns error if document does not exist in the DB
func Save(ctx context.Context, workflow *Tasks) error {
	if datastore.GetDB() == nil {
		return nil
	}
	if datastore.GetDB().Name() == "" {
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
	_, err := datastore.GetDBCollection().ReplaceOne(ctx, filter, workflow, opts)

	return err
}
