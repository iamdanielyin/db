package sqladapter

import (
	"context"
	"database/sql"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
)

type adapter struct {
	fns *AdapterFuncs
	ds  *db.DataSource
	db  *sql.DB
}

type AdapterFuncs struct {
	NativeCollectionNames    func(sqlhelper.SQLCommon, *db.DataSource) ([]string, error)
	NativeCollectionMetadata func(sqlhelper.SQLCommon, *db.DataSource) ([]schema.Metadata, error)
	Name                     func(sqlhelper.SQLCommon, *db.DataSource) string
	DriverName               func() string
}

func NewAdapter(fns *AdapterFuncs) *adapter {
	return &adapter{fns: fns}
}

func (a *adapter) NativeCollectionNames() ([]string, error) {
	if a.fns.NativeCollectionNames != nil {
		return a.fns.NativeCollectionNames(a.db, a.ds)
	}
	return nil, nil
}

func (a *adapter) NativeCollectionMetadata() ([]schema.Metadata, error) {
	if a.fns.NativeCollectionMetadata != nil {
		return a.fns.NativeCollectionMetadata(a.db, a.ds)
	}
	return nil, nil
}

func (a *adapter) Name() string {
	if a.fns.Name != nil {
		return a.fns.Name(a.db, a.ds)
	}
	return ""
}

func (a *adapter) DriverName() string {
	if a.fns.DriverName != nil {
		return a.fns.DriverName()
	}
	return ""
}

func (a *adapter) BeginTx() (db.Tx, error) {
	t, err := a.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	return &tx{a: a, ds: a.DataSource(), tx: t}, nil
}

func (a *adapter) DB() *sql.DB {
	return a.db
}

func (a *adapter) Model(s string) db.Collection {
	meta, has := db.Meta(s)
	if !has {
		return nil
	}
	return &model{
		a:      a,
		meta:   meta,
		common: a.db,
	}
}

func (a *adapter) Query(sql string, args ...interface{}) db.Query {
	rows, err := a.db.Query(sql, args...)
	return &query{rows: rows, err: err}
}

func (a *adapter) Exec(s string, args ...interface{}) (id interface{}, n uint64, err error) {
	var res sql.Result
	res, err = a.db.Exec(s, args...)
	if err != nil {
		return nil, 0, err
	}
	id, err = res.LastInsertId()
	if v, e := res.RowsAffected(); e != nil {
		return nil, 0, err
	} else {
		n = uint64(v)
	}
	return
}

func (a *adapter) DataSource() *db.DataSource {
	return a.ds
}

func (a *adapter) Open(source *db.DataSource) (db.Database, error) {
	ins, err := sql.Open(source.Adapter, source.DSN)
	if err != nil {
		return nil, err
	}
	if err := ins.Ping(); err != nil {
		return nil, err
	}
	return &adapter{db: ins, ds: source, fns: a.fns}, nil
}

func (a *adapter) Close() error {
	return a.db.Close()
}
