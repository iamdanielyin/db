package sqladapter

import (
	"database/sql"
	"github.com/yuyitech/db/pkg/db"
)

type tx struct {
	a  *adapter
	ds *db.DataSource
	tx *sql.Tx
}

func (t *tx) Name() string {
	if t.a.fns.Name != nil {
		return t.a.fns.Name(t.tx, t.ds)
	}
	return ""
}

func (t *tx) Model(s string) db.IModel {
	meta, has := db.Meta(s)
	if !has {
		return nil
	}
	return &model{
		a:      t.a,
		meta:   meta,
		common: t.tx,
	}
}

func (t *tx) Query(sql string, args ...interface{}) db.IQuery {
	rows, err := t.tx.Query(sql, args...)
	return &query{rows: rows, err: err}
}

func (t *tx) Exec(s string, args ...interface{}) (id interface{}, n uint64, err error) {
	var res sql.Result
	res, err = t.tx.Exec(s, args...)
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

func (t *tx) DataSource() *db.DataSource {
	return t.ds
}

func (t *tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *tx) Commit() error {
	return t.tx.Commit()
}
