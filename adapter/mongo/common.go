package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func nativeCollectionNames(db *mongo.Database) ([]string, error) {
	nameOnly := true
	names, err := db.ListCollectionNames(context.Background(), bson.D{}, &options.ListCollectionsOptions{
		NameOnly: &nameOnly,
	})
	return names, err
}
