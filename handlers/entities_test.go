package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/nanvenomous/discontent/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestFormToEntityMapper(t *testing.T) {
	tests := []struct {
		name          string
		entity        models.Article // Change type to models.Article
		expectedID    primitive.ObjectID
		expectedTitle string
	}{
		{
			name: "Valid data",
			entity: models.Article{
				ID:      primitive.NewObjectID(),
				Title:   "Test Title",
				Content: "Test Content",
				Author:  "Test Author",
			},
			expectedID:    primitive.NewObjectID(),
			expectedTitle: "Test Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with form data
			form := http.Request{Form: make(url.Values)}
			entityType := reflect.TypeOf(tt.entity)

			// Populate form data from the entity struct using reflection
			for i := 0; i < entityType.NumField(); i++ {
				field := entityType.Field(i)
				fieldValue := reflect.ValueOf(tt.entity).Field(i)

				if fieldValue.Kind() == reflect.String {
					form.Form[field.Name] = []string{fieldValue.String()}
				} else if fieldValue.Type() == reflect.TypeOf(primitive.ObjectID{}) {
					form.Form[field.Name] = []string{fieldValue.Interface().(primitive.ObjectID).Hex()}
				}
			}

			// Create a new entity value
			entityValue := reflect.New(entityType).Elem() // Change to reflect.New(entityType) instead of reflect.New(entityType.Elem())

			// Call the formToEntityMapper function
			_, err := formToEntityMapper(&form, entityValue)
			assert.Nil(t, err)

			fmt.Println(entityValue)
			// Validate the results
			entity := entityValue.Interface().(models.Article)
			assert.NotEmpty(t, entity.ID)
			assert.Equal(t, tt.expectedTitle, entity.Title)
		})
	}
}
