package mongo

import (
	"context"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func init() {
	db.RegisterAdapter(&adapter{})
}

const driverName = "mongo"

type adapter struct {
	ds            *db.DataSource
	cs            *connstring.ConnString
	client        *mongo.Client
	defaultDBName string
}

func (a *adapter) db() *mongo.Database {
	return a.client.Database(a.defaultDBName)
}

func (a *adapter) NativeCollectionNames() ([]string, error) {
	return nativeCollectionNames(a.db())
}

func (a *adapter) NativeCollectionMetadata() ([]schema.Metadata, error) {
	// MongoDB无本地结构
	return make([]schema.Metadata, 0), nil
}

func (a *adapter) Model(s string) db.Collection {
	meta, has := db.Meta(s)
	if !has {
		return nil
	}
	return &model{
		a:    a,
		meta: meta,
		db:   a.db(),
		coll: a.db().Collection(meta.NativeName),
	}
}

func (a *adapter) Name() string {
	return a.defaultDBName
}

func (a *adapter) Query(s string, i ...interface{}) db.Query {
	return &query{
		db:   a.db(),
		cmd:  s,
		args: i,
	}
}

func (a *adapter) Exec(s string, i ...interface{}) (interface{}, uint64, error) {
	// TODO 暂无可靠实现方式
	return nil, 0, nil
}

func (a *adapter) DriverName() string {
	return driverName
}

func (a *adapter) DataSource() *db.DataSource {
	return a.ds
}

func (a *adapter) Open(source *db.DataSource) (db.Database, error) {
	cs, err := connstring.Parse(source.DSN)
	if err != nil {
		return nil, err
	}
	clientOpts := options.Client().ApplyURI(source.DSN)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		return nil, err
	}
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, err
	}

	return &adapter{
		ds:            source,
		cs:            &cs,
		client:        client,
		defaultDBName: cs.Database,
	}, nil
}

func (a *adapter) Close() error {
	return a.client.Disconnect(context.Background())
}

func (a *adapter) BeginTx() (db.Tx, error) {
	sess, err := a.client.StartSession(options.Session())
	if err != nil {
		return nil, err
	}
	if err := sess.StartTransaction(); err != nil {
		return nil, err
	}
	return &tx{a: a, ds: a.DataSource(), sess: &sess}, nil
}
