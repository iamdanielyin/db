package sqladapter

import (
	"fmt"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/internal/templatex"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"reflect"
	"strings"
)

type model struct {
	a      *adapter
	common sqlhelper.SQLCommon
	meta   schema.Metadata
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

	var records []interface{}
	if itemV.Kind() == reflect.Array || itemV.Kind() == reflect.Slice {
		for i := 0; i < itemV.Len(); i++ {
			record := itemV.Index(i).Interface()
			records = append(records, record)
		}
	} else {
		records = append(records, i)
	}

	var (
		columns      []string
		placeholders []string
		values       []interface{}
	)
	for _, record := range records {
		ns, nv, err := sqlhelper.Map(record, &sqlhelper.MapOptions{
			IncludeZeroed: true,
			IncludeNil:    true,
		})
		var (
			fns []string
			fnv []interface{}
		)
		for i, n := range ns {
			field := m.meta.NativeFields()[n]
			if !field.IsAutoInc {
				fns = append(fns, n)
				fnv = append(fnv, nv[i])
			}
		}
		if err != nil {
			return nil, 0, err
		}
		if len(columns) == 0 {
			columns = fns
		}
		ps := make([]string, len(columns))
		for index := range columns {
			ps[index] = "?"
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(ps, ", ")))
		values = append(values, fnv...)
	}

	data := struct {
		Table   string
		Columns string
		Values  string
	}{
		Table:   m.meta.NativeName,
		Columns: strings.Join(columns, ", "),
		Values:  strings.Join(placeholders, ", "),
	}
	s, err := templatex.FastExecTXT(sqlhelper.InsertTemplate, data)
	if err != nil {
		return nil, 0, err
	}
	res, err := m.common.Exec(s, values...)
	if err != nil {
		return nil, 0, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return nil, 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, 0, err
	}
	return lastInsertId, uint64(rowsAffected), nil
}

func (m *model) Find(filter ...interface{}) db.FindResult {
	return newFindResult(m, filter)
}

func (m *model) Name() string {
	return m.meta.Name
}

func (m *model) Metadata() schema.Metadata {
	return m.meta
}

func (m *model) Database() db.Database {
	return m.a
}

func (m *model) Middleware() *db.AdapterMiddleware {
	return db.DefaultAdapterMiddleware
}
