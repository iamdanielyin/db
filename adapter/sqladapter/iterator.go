package sqladapter

import (
	"database/sql"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
)

type iterator struct {
	f    *findResult
	rows *sql.Rows
	err  error
}

func (i *iterator) Next(ptrToStruct interface{}) bool {
	next, err := sqlhelper.Next(i.rows, ptrToStruct)
	if err != nil {
		i.err = err
	}
	if i.f != nil {
		i.f.populateOne(ptrToStruct)
	}
	return next
}

func (i *iterator) Err() error {
	return i.err
}

func (i *iterator) Close() error {
	if i.rows != nil {
		return i.rows.Close()
	}
	return nil
}
