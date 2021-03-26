package mongo

import (
	"github.com/yuyitech/db/pkg/db"
	"testing"
)

func init() {
	_ = db.Connect(&db.DataSource{
		Name:      "tm",
		Adapter:   "mongo",
		DSN:       "mongodb://test:123456@127.0.0.1:27017/db?authSource=admin&connectTimeoutMS=300000&maxPoolSize=50",
		IsDefault: true,
	})
}

func TestAdapter_NativeCollectionMetadata(t *testing.T) {
	meta, err := db.DB("tm").NativeCollectionMetadata()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(meta)
}

func Test_adapter_NativeCollectionNames(t *testing.T) {
	names, err := db.DB().NativeCollectionNames()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(names)
}
