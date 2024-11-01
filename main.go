package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
	mu         sync.Mutex
)

type Entity struct {
	ID   string                 `json:"id" bson:"_id,omitempty"`
	Data map[string]interface{} `json:"data" bson:"data"`
}

func init() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	collection = client.Database("cms").Collection("entities")
}

func registerEntity(entity Entity) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := collection.InsertOne(context.TODO(), entity)
	return err
}

func getEntities(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var entities []Entity
	if err = cursor.All(context.TODO(), &entities); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entities)
}

func main() {
	http.HandleFunc("/api/entities", getEntities)
	http.ListenAndServe(":8080", nil)
}
