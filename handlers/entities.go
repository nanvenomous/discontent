// Package handlers provides HTTP handlers for managing entities.
package handlers

import (
	"net/http"
	"reflect"
	"time"

	"github.com/nanvenomous/discontent/data"
	"github.com/nanvenomous/discontent/models"
	"github.com/nanvenomous/discontent/reflection"
	"github.com/nanvenomous/discontent/ui"
)

// HandleEntity handles requests for entities.
func HandleEntity(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		stctFlds []reflection.StructField
	)

	collectionName := r.URL.Path[len("/api/entities/"):]

	structure, exists := models.GetStructFromCollectionName(collectionName)
	if !exists {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		stctFlds, err = reflection.GetStructFields(structure)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err := ui.Page(ui.SubmitEntityForm(structure, stctFlds)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if r.Method == http.MethodPost {
		time.Sleep(time.Second)
		entityValue := reflect.New(reflect.TypeOf(structure)).Elem()

		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entityValue, err = reflection.FormToEntityMapper(r, entityValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entityPtr := entityValue.Addr().Interface()
		oid, err := data.Create(entityPtr, collectionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

		_, err = reflection.AddIDToEntity(entityPtr, oid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stctFlds, err = reflection.GetStructFields(entityValue.Interface())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = ui.Page(ui.SubmitEntityForm(entityValue, stctFlds)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
