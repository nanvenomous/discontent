// Package handlers provides HTTP handlers for managing entities.
package handlers

import (
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/nanvenomous/discontent/models"
	"github.com/nanvenomous/discontent/ui"
)

// HandleEntity handles requests for entities.
func HandleEntity(w http.ResponseWriter, r *http.Request) {
	collectionName := r.URL.Path[len("/api/entities/"):]

	structure, exists := models.GetStructFromCollectionName(collectionName)
	if !exists {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	log.Println("got here")

	if r.Method == http.MethodGet {
		entityType := reflect.TypeOf(structure)

		err := ui.Page(ui.SubmitEntityForm(structure, entityType)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if r.Method == http.MethodPost {
		log.Println("got here post")
		time.Sleep(time.Second * 2)
		entityValue := reflect.New(reflect.TypeOf(structure)).Elem()

		for key, values := range r.Form {
			if len(values) > 0 {
				field := entityValue.FieldByNameFunc(func(name string) bool {
					return name == key
				})
				if field.IsValid() && field.CanSet() {
					field.SetString(values[0])
				}
			}
		}

		// entity := entityValue.Interface()
		w.WriteHeader(http.StatusCreated)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
