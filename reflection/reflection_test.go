package reflection

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/nanvenomous/discontent/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetStringFromValue(t *testing.T) {
	objID := primitive.NewObjectID()
	article := models.Article{
		ID:      objID,
		Title:   "Test Title",
		Content: "Test Content",
		Author:  "Test Author",
	}

	entityValue := reflect.ValueOf(article)
	for i := 0; i < entityValue.NumField(); i++ {
		fieldName := entityValue.Type().Field(i).Name
		expectedValue := ""

		switch fieldName {
		case "ID":
			expectedValue = article.ID.Hex()
		case "Title":
			expectedValue = article.Title
		case "Content":
			expectedValue = article.Content
		case "Author":
			expectedValue = article.Author
		}

		result := GetStringFromValue(entityValue.Type().Field(i), entityValue.Field(i))
		assert.Equal(t, expectedValue, result)
	}
}

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
			form := http.Request{Form: make(url.Values)}
			entityType := reflect.TypeOf(tt.entity)

			for i := 0; i < entityType.NumField(); i++ {
				field := entityType.Field(i)
				fieldValue := reflect.ValueOf(tt.entity).Field(i)

				if fieldValue.Kind() == reflect.String {
					form.Form[field.Name] = []string{fieldValue.String()}
				} else if fieldValue.Type() == reflect.TypeOf(primitive.ObjectID{}) {
					form.Form[field.Name] = []string{fieldValue.Interface().(primitive.ObjectID).Hex()}
				}
			}

			entityValue := reflect.New(entityType).Elem()

			_, err := FormToEntityMapper(&form, entityValue)
			assert.Nil(t, err)

			entity := entityValue.Interface().(models.Article)
			assert.NotEmpty(t, entity.ID)
			assert.Equal(t, tt.expectedTitle, entity.Title)
		})
	}
}

func TestAddIDToEntity(t *testing.T) {
	tests := []struct {
		name          string
		entity        models.Article
		expectedTitle string
	}{
		{
			name: "Valid entity with ID",
			entity: models.Article{
				Title:   "Test Title",
				Content: "Test Content",
				Author:  "Test Author",
			},
			expectedTitle: "Test Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pointer to the entity as required by addIDToEntity
			entityPtr := &tt.entity

			// Create a new ObjectID
			oid := primitive.NewObjectID()

			// Call addIDToEntity with pointer
			entityWithID, err := AddIDToEntity(entityPtr, oid)
			assert.Nil(t, err)

			// Validate the results
			entity := entityWithID.(*models.Article)
			assert.Equal(t, oid, entity.ID)
			assert.Equal(t, tt.expectedTitle, entity.Title)
		})
	}
}
