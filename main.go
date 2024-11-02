package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/amalfra/etag/v3"
	"github.com/nanvenomous/discontent/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed build/*
var buildFS embed.FS

var (
	EmbeddedResources fs.FS
	client            *mongo.Client
	collection        *mongo.Collection
	mu                sync.Mutex
)

type Entity struct {
	ID   string                 `json:"id" bson:"_id,omitempty"`
	Data map[string]interface{} `json:"data" bson:"data"`
}

func init() {
	var err error

	EmbeddedResources, err = fs.Sub(buildFS, "build")
	if err != nil {
		panic(err)
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	collection = client.Database("cms").Collection("entities")
}

// func registerEntity(entity Entity) error {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	_, err := collection.InsertOne(context.TODO(), entity)
// 	return err
// }

// func getEntities(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	cursor, err := collection.Find(context.TODO(), bson.D{})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var entities []Entity
// 	if err = cursor.All(context.TODO(), &entities); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(entities)
// }

func getBundledFile(r *http.Request) ([]byte, error) {
	if len(r.URL.Path) > 500 {
		return []byte{}, fmt.Errorf("The filepath you entered: %s was suspiciously long", r.URL.Path)
	}

	var pth = r.URL.Path
	if strings.HasPrefix(pth, "/") {
		pth = strings.TrimPrefix(pth, "/")
	}

	fl, err := EmbeddedResources.Open(pth)
	if err != nil {
		return []byte{}, err
	}
	defer fl.Close()

	return io.ReadAll(fl)
}

func serveResourceCachedETag(w http.ResponseWriter, r *http.Request,
	fileCheck func(r *http.Request) ([]byte, error)) {
	fmt.Println(r.URL.Path)

	w.Header().Set("Cache-Control", "max-age=0")
	content, err := fileCheck(r)
	if err != nil {
		http.Error(w, "Could not locate the file to serve: "+err.Error(), http.StatusNotFound)
		return
	}

	etg := etag.Generate(string(content), false)
	if r.Header.Get("If-None-Match") == etg {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etg)
	http.ServeContent(w, r, r.URL.Path, time.Unix(0, 0), bytes.NewReader(content))
}

func main() {
	var err error
	mux := http.NewServeMux() // Create a new ServeMux

	mux.HandleFunc("/api/entities/", handlers.HandleEntity)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains([]string{"", "/"}, r.URL.Path) {
			// Handle root path if needed
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("status ok"))
			return
		} else {
			serveResourceCachedETag(w, r, getBundledFile)
		}
	})

	fmt.Println("listening on port :8080")
	err = http.ListenAndServe(":8080", mux) // Use the mux in ListenAndServe
	if err != nil {
		panic(err)
	}
}
