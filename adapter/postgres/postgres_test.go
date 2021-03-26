package postgres

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/yuyitech/db/pkg/db"
	"testing"
)

func init() {
	_ = db.Connect(&db.DataSource{
		Name:    "asm",
		Adapter: "postgres",
		DSN:     "host=127.0.0.1 port=5432 user=root dbname=test password=123456 sslmode=disable timezone='Asia/Shanghai'",
	})
}

func TestAdapter_NativeCollectionMetadata(t *testing.T) {
	meta, err := db.DB("asm").NativeCollectionMetadata()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(meta)
}
