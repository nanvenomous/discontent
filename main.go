// Package main is the entry point for the Discontent CMS.
package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
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
	embeddedResources fs.FS
	client            *mongo.Client
	collection        *mongo.Collection
	mu                sync.Mutex
)

// type Entity struct {
// 	ID   string                 `json:"id" bson:"_id,omitempty"`
// 	Data map[string]interface{} `json:"data" bson:"data"`
// }

func init() {
	var err error

	embeddedResources, err = fs.Sub(buildFS, "build")
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
		return []byte{}, fmt.Errorf(
			"The filepath you entered: %s was suspiciously long",
			r.URL.Path,
		)
	}

	var pth = r.URL.Path
	if strings.HasPrefix(pth, "/") {
		pth = strings.TrimPrefix(pth, "/")
	}

	fl, err := embeddedResources.Open(pth)
	if err != nil {
		return []byte{}, err
	}
	defer fl.Close()

	return io.ReadAll(fl)
}

func serveResourceCachedETag(w http.ResponseWriter, r *http.Request,
	fileCheck func(r *http.Request) ([]byte, error),
) {
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"[%s] %s %s",
			r.Method,
			r.URL.String(),
			r.URL.Query().Encode(),
		)
		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error
	mux := http.NewServeMux()

	mux.HandleFunc("/api/entities/", handlers.HandleEntity)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains([]string{"", "/"}, r.URL.Path) {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("status ok"))
			if err != nil {
				http.Error(
					w,
					"Could not write the status message: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			return
		}
		serveResourceCachedETag(w, r, getBundledFile)
	})

	// Apply logging middleware
	loggedMux := loggingMiddleware(mux)

	log.Println("listening on port :8080")
	err = http.ListenAndServe(":8080", loggedMux)
	if err != nil {
		panic(err)
	}
}
