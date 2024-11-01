package handlers

import (
	"net/http"
	"reflect"

	"github.com/nanvenomous/discontent/examples"
	"github.com/nanvenomous/discontent/ui"
)

func HandleEntity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		collectionName := r.URL.Path[len("/api/entities/"):]

		structure, exists := examples.GetStructFromCollectionName(collectionName)
		if !exists {
			http.Error(w, "Collection not found", http.StatusNotFound)
			return
		}
		entityType := reflect.TypeOf(structure)

		err := ui.Page(ui.SubmitEntityForm(entityType)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if r.Method == http.MethodPost {
		collectionName := r.URL.Path[len("/api/entities/"):]

		structure, exists := examples.GetStructFromCollectionName(collectionName)
		if !exists {
			http.Error(w, "Collection not found", http.StatusNotFound)
			return
		}

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
