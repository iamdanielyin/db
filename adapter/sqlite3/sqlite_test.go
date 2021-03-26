package sqlite3

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/yuyitech/db/pkg/db"
	"testing"
)

func init() {
	_ = db.Connect(&db.DataSource{
		Name:    "drone",
		Adapter: "sqlite3",
		DSN:     "file:///Users/yinfxs/gopath/src/github.com/yinfxs/test-kuu2/database.sqlite",
	})
}

func TestAdapter_NativeCollectionMetadata(t *testing.T) {
	meta, err := db.DB("drone").NativeCollectionMetadata()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(meta)
}
