package sqlhelper

import (
	"database/sql"
	"github.com/yuyitech/db/internal/reflectx"
	"github.com/yuyitech/db/pkg/db"
	"reflect"
)

var (
	mapper = reflectx.NewMapper("json")
)

func One(rows *sql.Rows, dst interface{}) error {
	return fetchRow(rows, dst)
}

func All(rows *sql.Rows, dst interface{}) error {
	return fetchRows(rows, dst)
}

func Next(rows *sql.Rows, dst interface{}) (bool, error) {
	return next(rows, dst)
}

func next(rows *sql.Rows, dst interface{}) (bool, error) {
	if rows == nil {
		return false, db.ErrNoMoreRows
	}

	var columns []string
	var err error

	dstv := reflect.ValueOf(dst)

	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return false, db.ErrExpectingPointer
	}

	itemV := dstv.Elem()

	if columns, err = rows.Columns(); err != nil {
		return false, err
	}

	reset(dst)

	next := rows.Next()

	if !next {
		if err = rows.Err(); err != nil {
			return next, err
		}
		return next, db.ErrNoMoreRows
	}

	itemT := itemV.Type()
	item, err := fetchResult(rows, itemT, columns)

	if err != nil {
		return next, err
	}

	if itemT.Kind() == reflect.Ptr {
		itemV.Set(item)
	} else {
		itemV.Set(reflect.Indirect(item))
	}

	return next, nil
}

// fetchRow receives a *sql.Rows value and tries to map all the rows into a
// single struct given by the pointer `dst`.
func fetchRow(rows *sql.Rows, dst interface{}) error {
	defer func() { _ = rows.Close() }()
	_, err := next(rows, dst)

	if err == db.ErrNoMoreRows {
		return db.ErrRecordNotFound
	}

	return err
}

// fetchRows receives a *sql.Rows value and tries to map all the rows into a
// slice of struct given by the pointer `dst`.
func fetchRows(rows *sql.Rows, dst interface{}) error {
	var err error
	defer func() { _ = rows.Close() }()

	// Destination.
	dstv := reflect.ValueOf(dst)

	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return db.ErrExpectingPointer
	}

	if dstv.Elem().Kind() != reflect.Slice {
		return db.ErrExpectingSlicePointer
	}

	if dstv.Kind() != reflect.Ptr || dstv.Elem().Kind() != reflect.Slice || dstv.IsNil() {
		return db.ErrExpectingSliceMapStruct
	}

	var columns []string
	if columns, err = rows.Columns(); err != nil {
		return err
	}

	slicev := dstv.Elem()
	itemT := slicev.Type().Elem()

	reset(dst)

	var rowsAffected int
	for rows.Next() {
		rowsAffected++

		item, err := fetchResult(rows, itemT, columns)
		if err != nil {
			return err
		}
		if itemT.Kind() == reflect.Ptr {
			slicev = reflect.Append(slicev, item)
		} else {
			slicev = reflect.Append(slicev, reflect.Indirect(item))
		}
	}

	dstv.Elem().Set(slicev)

	if err := rows.Err(); err != nil {
		return err
	} else if rowsAffected == 0 {
		return db.ErrRecordNotFound
	}

	return rows.Err()
}

func fetchResult(rows *sql.Rows, itemT reflect.Type, columns []string) (reflect.Value, error) {
	var item reflect.Value
	var err error

	objT := itemT

	switch objT.Kind() {
	case reflect.Map:
		item = reflect.MakeMap(objT)
	case reflect.Struct:
		item = reflect.New(objT)
	case reflect.Ptr:
		objT = itemT.Elem()
		if objT.Kind() != reflect.Struct {
			return item, db.ErrExpectingMapOrStruct
		}
		item = reflect.New(objT)
	default:
		return item, db.ErrExpectingMapOrStruct
	}

	switch objT.Kind() {
	case reflect.Struct:

		values := make([]interface{}, len(columns))
		typeMap := mapper.TypeMap(itemT)
		fieldMap := typeMap.Names

		for i, k := range columns {
			fi, ok := fieldMap[k]
			if !ok {
				values[i] = new(interface{})
				continue
			}

			f := reflectx.FieldByIndexes(item, fi.Index)
			values[i] = f.Addr().Interface()
		}

		if err = rows.Scan(values...); err != nil {
			return item, err
		}
	case reflect.Map:

		columns, err := rows.Columns()
		if err != nil {
			return item, err
		}

		values := make([]interface{}, len(columns))
		for i := range values {
			if itemT.Elem().Kind() == reflect.Interface {
				values[i] = new(interface{})
			} else {
				values[i] = reflect.New(itemT.Elem()).Interface()
			}
		}

		if err = rows.Scan(values...); err != nil {
			return item, err
		}

		for i, column := range columns {
			item.SetMapIndex(reflect.ValueOf(column), reflect.Indirect(reflect.ValueOf(values[i])))
		}
	}

	return item, nil
}

func reset(data interface{}) {
	// Resetting element.
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var z reflect.Value

	switch v.Kind() {
	case reflect.Slice:
		z = reflect.MakeSlice(t, 0, v.Cap())
	default:
		z = reflect.Zero(t)
	}

	v.Set(z)
}
