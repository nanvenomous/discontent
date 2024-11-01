package examples

type Article struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	Title   string `json:"title" bson:"title"`
	Content string `json:"content" bson:"content"`
	Author  string `json:"author" bson:"author"`
}

type Category struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

type Comment struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	ArticleID string `json:"article_id" bson:"article_id"`
	Content   string `json:"content" bson:"content"`
	Author    string `json:"author" bson:"author"`
}

var collectionMap = map[string]interface{}{
	"articles":   Article{},
	"categories": Category{},
	"comments":   Comment{},
}

var structMap = map[interface{}]string{
	Article{}:  "articles",
	Category{}: "categories",
	Comment{}:  "comments",
}

func GetStructFromCollectionName(collectionName string) (interface{}, bool) {
	structure, exists := collectionMap[collectionName]
	return structure, exists
}

func GetCollectionNameFromStruct(entity interface{}) (string, bool) {
	collectionName, exists := structMap[entity]
	return collectionName, exists
}
