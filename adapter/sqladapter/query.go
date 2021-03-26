package sqladapter

import (
	"database/sql"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
)

type query struct {
	f    *findResult
	rows *sql.Rows
	err  error
}

func (q *query) Iterator() (db.Iterator, error) {
	if q.err != nil {
		return nil, q.err
	}
	return &iterator{
		f:    q.f,
		rows: q.rows,
		err:  q.err,
	}, q.err
}

func (q *query) One(ptrToStruct interface{}) error {
	if err := sqlhelper.One(q.rows, ptrToStruct); err != nil {
		q.err = err
	}
	return q.err
}

func (q *query) All(sliceOfStruct interface{}) error {
	if err := sqlhelper.All(q.rows, sliceOfStruct); err != nil {
		q.err = err
	}
	return q.err
}
