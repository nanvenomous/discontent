// Package data contains the data for the application.
package data

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

// Setup initializes the MongoDB client and collection.
func Setup() error {
	var err error

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}
	db = client.Database("cms")
	return nil
}

// Create saves a new entity to the specified collection in MongoDB
func Create[T any](entity T, collectionName string) (primitive.ObjectID, error) {
	collection := db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return primitive.ObjectID{}, fmt.Errorf(
			"error creating entity in collection %s: %w",
			collectionName,
			err,
		)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("inserted ID is not of type primitive.ObjectID")
	}

	return id, nil
}
