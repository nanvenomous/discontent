// Package models defines the data structures for the Discontent CMS.
package models

// Article represents a blog article.
type Article struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	Title   string `json:"title" bson:"title"`
	Content string `json:"content" bson:"content"`
	Author  string `json:"author" bson:"author"`
}

// Category represents a category.
type Category struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

// Comment represents a comment on an article.
type Comment struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	ArticleID string `json:"article_id" bson:"article_id"`
	Content   string `json:"content" bson:"content"`
	Author    string `json:"author" bson:"author"`
}

// collectionMap maps collection names to their corresponding struct types.
var collectionMap = map[string]any{
	"articles":   Article{},
	"categories": Category{},
	"comments":   Comment{},
}

// structMap maps struct types to their corresponding collection names.
var structMap = map[any]string{
	Article{}:  "articles",
	Category{}: "categories",
	Comment{}:  "comments",
}

// GetStructFromCollectionName retrieves the struct associated with the given collection name.
func GetStructFromCollectionName(collectionName string) (any, bool) {
	structure, exists := collectionMap[collectionName]
	return structure, exists
}

// GetCollectionNameFromStruct retrieves the collection name associated with the given struct.
func GetCollectionNameFromStruct(entity any) (string, bool) {
	collectionName, exists := structMap[entity]
	return collectionName, exists
}
