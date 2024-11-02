// Package reflection contains helper functions for reflection
package reflection

import (
	"errors"
	"net/http"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetStringFromValue gets the string value from a reflect.StructField
func GetStringFromValue(fld reflect.StructField, val reflect.Value) string {
	// Handle ObjectID type
	if fld.Type.Name() == "ObjectID" {
		if val.Interface() != primitive.NilObjectID {
			return val.Interface().(primitive.ObjectID).Hex()
		}
		return ""
	}

	// Handle string type
	if fld.Type.Kind() == reflect.String {
		return val.String()
	}

	return ""
}

// FormToEntityMapper adds the form data to the entity struct
func FormToEntityMapper(r *http.Request, entityValue reflect.Value) (reflect.Value, error) {
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

// AddIDToEntity adds the objectID to entity.ID field
func AddIDToEntity(entity any, oid primitive.ObjectID) (any, error) {
	val := reflect.ValueOf(entity)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil, errors.New("entity must be a non-nil pointer")
	}

	idField := val.Elem().FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() {
		return nil, errors.New("entity does not have a valid ID field")
	}

	idField.Set(reflect.ValueOf(oid))
	return entity, nil
}
