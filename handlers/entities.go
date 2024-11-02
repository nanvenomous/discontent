// Package handlers provides HTTP handlers for managing entities.
package handlers

import (
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/nanvenomous/discontent/data"
	"github.com/nanvenomous/discontent/models"
	"github.com/nanvenomous/discontent/ui"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func formToEntityMapper(r *http.Request, entityValue reflect.Value) (reflect.Value, error) {
	for key, values := range r.Form {
		if len(values) > 0 {
			field := entityValue.FieldByNameFunc(func(name string) bool {
				return name == key
			})
			if field.IsValid() && field.CanSet() {
				if field.Type().Name() == "ObjectID" {
					if values[0] != "" {
						objID, err := primitive.ObjectIDFromHex(values[0])
						if err != nil {
							return entityValue, err
						}
						field.Set(reflect.ValueOf(objID))
					}
				} else {
					field.SetString(values[0])
				}
			}
		}
	}
	return entityValue, nil
}

// HandleEntity handles requests for entities.
func HandleEntity(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	collectionName := r.URL.Path[len("/api/entities/"):]

	structure, exists := models.GetStructFromCollectionName(collectionName)
	if !exists {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	entityType := reflect.TypeOf(structure)

	if r.Method == http.MethodGet {
		err := ui.Page(ui.SubmitEntityForm(structure, entityType)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if r.Method == http.MethodPost {
		time.Sleep(time.Second * 4)
		entityValue := reflect.New(reflect.TypeOf(structure)).Elem()

		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println(r.Form)

		entityValue, err = formToEntityMapper(r, entityValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println(entityValue)

		entity := entityValue.Interface()
		oid, err := data.Create(entity, collectionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		entityWithID, ok := entity.(interface{ SetID(primitive.ObjectID) })
		if !ok {
			http.Error(w, "Entity does not support setting ID", http.StatusInternalServerError)
			return
		}
		entityWithID.SetID(oid)
		w.WriteHeader(http.StatusCreated)

		err = ui.Page(ui.SubmitEntityForm(entity, entityType)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
