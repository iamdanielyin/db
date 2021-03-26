package mongo

import (
	"context"
	"github.com/yuyitech/db/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type tx struct {
	a    *adapter
	ds   *db.DataSource
	sess *mongo.Session
}

func (t *tx) db() *mongo.Database {
	return (*t.sess).Client().Database(t.a.defaultDBName)
}

func (t *tx) Model(s string) db.IModel {
	meta, has := db.Meta(s)
	if !has {
		return nil
	}
	return &model{
		a:    t.a,
		meta: meta,
		db:   t.db(),
		coll: t.db().Collection(meta.NativeName),
	}
}

func (t *tx) Name() string {
	return t.a.defaultDBName
}

func (t *tx) Query(s string, i ...interface{}) db.IQuery {
	return &query{
		db:   t.db(),
		cmd:  s,
		args: i,
	}
}

func (t *tx) Exec(s string, i ...interface{}) (interface{}, uint64, error) {
	// TODO 暂无可靠实现方式
	return nil, 0, nil
}

func (t *tx) DriverName() string {
	return driverName
}

func (t *tx) DataSource() *db.DataSource {
	return t.ds
}

func (t *tx) Rollback() error {
	defer (*t.sess).EndSession(context.Background())
	return (*t.sess).AbortTransaction(context.Background())
}

func (t *tx) Commit() error {
	defer (*t.sess).EndSession(context.Background())
	return (*t.sess).CommitTransaction(context.Background())
}
