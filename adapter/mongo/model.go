package mongo

import (
	"context"
	"github.com/yuyitech/db/internal/reflectx"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type model struct {
	a    *adapter
	db   *mongo.Database
	coll *mongo.Collection
	meta schema.Metadata
}

func (m *model) Name() string {
	return m.meta.Name
}

func (m *model) Metadata() schema.Metadata {
	return m.meta
}

func (m *model) Create(i interface{}) (interface{}, uint64, error) {
	itemV := reflect.ValueOf(i)
	if !itemV.IsValid() {
		return nil, 0, nil
	}
	itemT := itemV.Type()
	if itemT.Kind() == reflect.Ptr {
		i = itemV.Elem().Interface()
		itemV = reflect.ValueOf(i)
		itemT = itemV.Type()
	}

	if itemV.Kind() == reflect.Array || itemV.Kind() == reflect.Slice {
		res, err := m.coll.InsertMany(context.Background(), reflectx.ToInterfaceArray(i))
		if err != nil {
			return nil, 0, err
		}
		if len(res.InsertedIDs) > 0 {
			return res.InsertedIDs[0], uint64(len(res.InsertedIDs)), nil
		}
		return nil, 0, nil
	} else {
		res, err := m.coll.InsertOne(context.Background(), i)
		if err != nil {
			return nil, 0, err
		}
		var n uint64
		if res.InsertedID != nil {
			n = 1
		}
		return res.InsertedID, n, nil
	}
}

func (m *model) Find(filter ...interface{}) db.FindResult {
	return newFindResult(m, filter)
}

func (m *model) Database() db.Database {
	return m.a
}

func (m *model) Middleware() *db.AdapterMiddleware {
	return db.DefaultAdapterMiddleware
}
