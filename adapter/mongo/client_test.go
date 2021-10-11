package mongo

import (
	"github.com/yuyitech/db"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	_, err := db.Connect(db.DataSource{
		Name:    "test",
		Adapter: "mongo",
		URI:     uri,
	})
	if err != nil {
		t.Fatal(err)
	}
}
