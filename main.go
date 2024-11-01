package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/nanvenomous/discontent/handlers"
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
	var err error
	http.HandleFunc("/api/entities", getEntities)
	http.HandleFunc("/api/entities/", handlers.HandleEntity)

	fmt.Println("listening on port :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
