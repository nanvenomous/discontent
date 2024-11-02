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
	var (
		err error
		mux = http.NewServeMux()
	)

	embeddedResources, err = fs.Sub(buildFS, "build")
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
	}
}
