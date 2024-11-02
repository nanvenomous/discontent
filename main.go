// Package main is the entry point for the Discontent CMS.
package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/amalfra/etag/v3"
	"github.com/nanvenomous/discontent/data"
	"github.com/nanvenomous/discontent/handlers"
)

//go:embed build/*
var buildFS embed.FS

var (
	embeddedResources fs.FS
)

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
		var color string
		switch r.Method {
		case http.MethodGet:
			color = "\033[32m" // Green for GET
		case http.MethodPost:
			color = "\033[34m" // Blue for POST
		case http.MethodPut:
			color = "\033[33m" // Yellow for PUT
		case http.MethodDelete:
			color = "\033[31m" // Red for DELETE
		default:
			color = "\033[0m" // Default color for other methods
		}
		log.Printf(
			"%s[%s]%s %s %s",
			color,
			r.Method,
			"\033[0m", // Reset color
			r.URL.String(),
			r.URL.Query().Encode(),
		)
		next.ServeHTTP(w, r)
	})
}

func run() error {
	var (
		err error
		mux = http.NewServeMux()
	)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	embeddedResources, err = fs.Sub(buildFS, "build")
	if err != nil {
		return err
	}

	err = data.Setup()
	if err != nil {
		return err
	}

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
	return err
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
