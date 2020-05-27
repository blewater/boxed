package datastore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// create a workflow document
func create(ctx context.Context, workflow DBDocument) error {
	if workflow.GetID() == primitive.NilObjectID {
		if err := workflow.SetID(primitive.NewObjectID().String()); err != nil {
			return err
		}
	}

	workflow.SetCreatedAt(time.Now())

	_, err := GetDBCollection().InsertOne(ctx, workflow)

	return err
}

// read returns a document based on the filter
func read(ctx context.Context, filter bson.M) (DBDocument, error) {
	var do DBDocument

	res := GetDBCollection().FindOne(ctx, filter)
	err := res.Decode(&do)

	return do, err
}

// LoadByName returns the database organization workflow database
func LoadByName(ctx context.Context, workflowNameKey string) (interface{}, error) {
	filter := bson.M{"workflowNameKey": workflowNameKey}

	dbWorkflow, err := read(ctx, filter)

	if err != nil && err.Error() != "mongo: no documents in result" {
		return nil, err
	}

	if dbWorkflow == nil {
		return nil, errors.New("failed to load workflow from mongo or id found to be nil")
	}

	if dbWorkflow.GetID() == primitive.NilObjectID {
		return nil, fmt.Errorf("workflow from mongo with nil id: %s", dbWorkflow.GetID())
	}

	return dbWorkflow, nil
}

// Upsert workflow tasks collection to initialized mongo instance.
// If the ID is a nil object then proceed with create
// otherwise with replace or update of the document collection
// returns database non-availability errors
func Upsert(ctx context.Context, workflow DBDocument) error {
	if GetDB() == nil {
		return nil
	}
	if GetDB().Name() == "" {
		return fmt.Errorf("database error")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	workflow.SetUpdatedAt(time.Now())

	// Create first if not existing
	if workflow.GetID() == primitive.NilObjectID {
		return create(ctx, workflow)
	}

	filter := bson.M{"_id": workflow.GetID()}

	opts := options.Replace().SetUpsert(false)
	_, err := GetDBCollection().ReplaceOne(ctx, filter, workflow, opts)

	return err
}
