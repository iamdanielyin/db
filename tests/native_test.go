package tests

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

type Member struct {
	FirstName string
	LastName  string
}

func TestNative(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		t.Fatal(err)
	}
	collection := client.Database("hello").Collection("member")
	var docs = []interface{}{
		Member{
			FirstName: "Eason",
			LastName:  "Chan",
		},
		Member{
			FirstName: "Daniel",
			LastName:  "Wu",
		},
		Member{
			FirstName: "Steve",
			LastName:  "Jobs",
		},
	}
	_, _ = collection.InsertMany(context.Background(), docs)
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		t.Fatal(err)
	}
	var results []Member
	if err = cur.All(context.Background(), &results); err != nil {
		t.Fatal(err)
	}
	t.Log(results)
}
