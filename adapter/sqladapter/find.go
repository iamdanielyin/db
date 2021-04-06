package sqladapter

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/internal/reflectx"
	"github.com/yuyitech/db/internal/templatex"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"math"
	"reflect"
	"strings"
)

type findResult struct {
	m      *model
	a      *adapter
	common sqlhelper.SQLCommon
	meta   schema.Metadata

	rows *sql.Rows

	err       error
	page      uint
	size      uint
	filters   []interface{}
	orders    map[string]bool
	selects   map[string]bool
	populates map[string]*db.PopulateOptions
}

type findTemplateData struct {
	Distinct string
	Columns  string
	Table    string
	Where    string
	Joins    string
	GroupBy  string
	OrderBy  string
	Limit    uint
	Offset   uint
}

func newFindResult(m *model, filter []interface{}) *findResult {
	result := &findResult{
		m:      m,
		a:      m.a,
		common: m.common,
		meta:   m.meta,

		orders:    make(map[string]bool),
		selects:   make(map[string]bool),
		populates: make(map[string]*db.PopulateOptions),
	}
	if len(filter) > 0 {
		result.filters = append(result.filters, filter...)
	}
	return result
}

func where(filters []interface{}, meta schema.Metadata) (string, []interface{}) {
	var (
		buffer bytes.Buffer
		attrs  []interface{}
	)
	for i, filter := range filters {
		cmp := sqlhelper.ParseFilter(filter, func(cmp *db.Comparison) {
			field := meta.Fields[cmp.Key]
			cmp.Key = field.MustNativeName()
		})
		s := cmp.CombineStmts()
		if s == "" {
			continue
		}
		if len(filters) > 1 {
			buffer.WriteString(fmt.Sprintf("(%s)", s))
		} else {
			buffer.WriteString(fmt.Sprintf("%s", s))
		}
		if i < len(filters)-1 {
			buffer.WriteString(" AND ")
		}
		attrs = append(attrs, cmp.Args...)
	}
	return buffer.String(), attrs
}

func orderBy(orders map[string]bool, meta schema.Metadata) string {
	var (
		buffer bytes.Buffer
		i      int
	)
	for k, v := range orders {
		field := meta.Fields[k]
		nativeName := k
		if field.NativeName != "" {
			nativeName = field.NativeName
		}
		if v {
			nativeName = fmt.Sprintf("%s DESC", nativeName)
		}
		buffer.WriteString(nativeName)
		if i < len(orders)-1 {
			buffer.WriteString(", ")
		}
		i++
	}
	return buffer.String()
}

func pagination(page, size uint) (offset uint, limit uint) {
	if size > 0 && page > 0 {
		offset = page - 1
		limit = size
	}
	return
}

func columns(selects map[string]bool, meta schema.Metadata) string {
	var (
		buffer bytes.Buffer
		i      int
	)
	for k := range selects {
		field := meta.Fields[k]
		nativeName := field.MustNativeName()
		buffer.WriteString(nativeName)
		if i < len(selects)-1 {
			buffer.WriteString(", ")
		}
		i++
	}
	return buffer.String()
}

func (f *findResult) executeQuery(queryTemplate string, fn func(*findTemplateData)) (*sql.Rows, error) {
	columns := columns(f.selects, f.meta)
	orderBy := orderBy(f.orders, f.meta)
	offset, limit := pagination(f.page, f.size)
	where, attrs := where(f.filters, f.meta)
	data := findTemplateData{
		Table:   f.meta.NativeName,
		Columns: columns,
		Where:   where,
		OrderBy: orderBy,
		Limit:   limit,
		Offset:  offset,
	}
	if fn != nil {
		fn(&data)
	}
	s, err := templatex.FastExecTXT(queryTemplate, data)
	if err != nil {
		f.err = err
		return nil, err
	}
	rows, err := f.common.Query(s, attrs...)
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func (f *findResult) populateOne(ptrToStruct interface{}) {
	dstReflectValue := reflect.ValueOf(ptrToStruct)
	mapper := reflectx.NewMapper("json")

	if len(f.populates) > 0 {
		for path, opts := range f.populates {
			field := f.meta.Fields[path]
			if field.MustName() == "" {
				continue
			}
			if opts == nil {
				opts = new(db.PopulateOptions)
			}
			if opts.Model == nil {
				opts.Model = db.Model(field.RelationshipModel)
			}
			if opts.Model == nil {
				continue
			}

			var (
				fieldValue = mapper.FieldByName(dstReflectValue, field.NativeName)
				v          = reflect.New(fieldValue.Type())
				d          = reflect.Indirect(v).Addr().Interface()
			)

			localField := f.meta.Fields[field.RelationshipLocalField]
			res := opts.Model.Find(opts.Match).Select(opts.Select...).Where(db.Cond{
				field.RelationshipForeignField: mapper.FieldByName(dstReflectValue, localField.NativeName).Interface(),
			})
			switch field.RelationshipKind {
			case db.RelationshipHasOne:
				f.err = res.One(d)
			case db.RelationshipHasMany:
				f.err = res.All(d)
			case db.RelationshipRefOne:
				f.err = res.One(d)
			case db.RelationshipRefMany:
				f.err = res.All(d)
			}
			if f.err != nil {
				continue
			}
			if fieldValue.CanAddr() {
				fieldValue.Set(v.Elem())
			} else {
				fieldValue.Set(v)
			}
		}
	}
}

func (f *findResult) One(ptrToStruct interface{}) error {
	rows, err := f.executeQuery(sqlhelper.SelectTemplate, func(data *findTemplateData) {
		data.Limit = 1
		data.Offset = 0
	})
	if err != nil {
		f.err = err
		return err
	}
	if err := sqlhelper.One(rows, ptrToStruct); err != nil {
		f.err = err
		return err
	}
	f.populateOne(ptrToStruct)
	return f.err
}

func (f *findResult) populateAll(sliceOfStruct interface{}) {
	mapper := reflectx.NewMapper("json")
	dstReflectValue := reflect.ValueOf(sliceOfStruct)
	dstIndirectValue := reflect.Indirect(dstReflectValue)

	if len(f.populates) > 0 && dstIndirectValue.Len() > 0 {
		for path, opts := range f.populates {
			populateField := f.meta.Fields[path]
			if populateField.MustName() == "" {
				continue
			}
			if opts == nil {
				opts = new(db.PopulateOptions)
			}
			if opts.Model == nil {
				opts.Model = db.Model(populateField.RelationshipModel)
			}
			if opts.Model == nil {
				continue
			}

			var (
				localField            = f.meta.Fields[populateField.RelationshipLocalField]
				localFieldValues      []interface{}
				localFieldElementsMap = make(map[interface{}][]int)
			)

			for i := 0; i < dstIndirectValue.Len(); i++ {
				rv := mapper.FieldByName(dstIndirectValue.Index(i), localField.NativeName)
				value := rv.Interface()
				localFieldValues = append(localFieldValues, value)
				localFieldElementsMap[value] = append(localFieldElementsMap[value], i)
			}

			var (
				foreignDstType     = mapper.FieldByName(dstIndirectValue.Index(0), populateField.NativeName).Type()
				foreignDstTypeKind = foreignDstType.Kind()
				foreignDstValue    reflect.Value
			)

			if foreignDstTypeKind == reflect.Array || foreignDstTypeKind == reflect.Slice {
				foreignDstValue = reflect.New(foreignDstType)
			} else {
				foreignDstValue = reflect.New(reflect.SliceOf(foreignDstType))
			}
			foreignDst := foreignDstValue.Interface()

			res := opts.Model.Find(opts.Match).Select(opts.Select...).Where(db.Cond{
				fmt.Sprintf("%s in", populateField.RelationshipForeignField): localFieldValues,
			})
			f.err = res.All(foreignDst)

			if f.err != nil {
				continue
			}

			var (
				foreignDstElem = foreignDstValue.Elem()
				foreignDstKind = foreignDstElem.Kind()
				dstElemInit    = make(map[int]bool)
			)

			setPopulateFieldValue := func(foreignDstItem reflect.Value) {
				foreignMetaField := opts.Model.Metadata().Fields[populateField.RelationshipForeignField] // 引用档案的外键字段
				foreignFieldValue := mapper.FieldByName(foreignDstItem, foreignMetaField.NativeName)     // 外键字段的值

				effectedElements := localFieldElementsMap[foreignFieldValue.Interface()]
				for _, index := range effectedElements {
					populateFieldValue := mapper.FieldByName(dstIndirectValue.Index(index), populateField.NativeName)

					var values reflect.Value
					if dstElemInit[index] {
						values = populateFieldValue
					} else {
						values = reflect.New(populateFieldValue.Type()).Elem()
						dstElemInit[index] = true
					}

					valuesKind := values.Kind()
					switch valuesKind {
					case reflect.Array, reflect.Slice:
						values = reflect.Append(values, foreignDstItem)
					case reflect.Struct:
						values = foreignDstItem
					}

					if populateFieldValue.CanAddr() {
						populateFieldValue.Set(values)
					} else {
						populateFieldValue.Set(values.Addr())
					}
				}
			}

			switch foreignDstKind {
			case reflect.Struct:
				setPopulateFieldValue(foreignDstElem)
			case reflect.Array, reflect.Slice:
				for i := 0; i < foreignDstElem.Len(); i++ { // 循环引用档案的每一行
					setPopulateFieldValue(foreignDstElem.Index(i))
				}
			}
		}
	}
}

func (f *findResult) All(sliceOfStruct interface{}) error {
	rows, err := f.executeQuery(sqlhelper.SelectTemplate, nil)
	if err != nil {
		f.err = err
		return err
	}
	if err := sqlhelper.All(rows, sliceOfStruct); err != db.ErrRecordNotFound {
		f.err = err
		return err
	}

	f.populateAll(sliceOfStruct)
	return f.err
}

func (f *findResult) Iterator() (db.Iterator, error) {
	rows, err := f.executeQuery(sqlhelper.SelectTemplate, nil)
	if err != nil {
		return nil, err
	}
	return &iterator{
		f:    f,
		rows: rows,
		err:  err,
	}, err
}

func (f *findResult) Page(u uint) db.FindResult {
	f.page = u
	return f
}

func (f *findResult) Size(u uint) db.FindResult {
	f.size = u
	return f
}

func (f *findResult) Order(s ...string) db.FindResult {
	for _, item := range s {
		var descend bool
		if strings.HasPrefix(item, "-") {
			descend = true
			item = item[1:]
		}
		f.orders[item] = descend
	}
	return f
}

func (f *findResult) Select(s ...string) db.FindResult {
	for _, item := range s {
		f.selects[item] = true
	}
	return f
}

func (f *findResult) Where(filter interface{}) db.FindResult {
	f.filters = append(f.filters, filter)
	return f
}

func (f *findResult) And(filters ...db.Cond) db.FindResult {
	return f.Where(db.And(filters...))
}

func (f *findResult) Or(filters ...db.Cond) db.FindResult {
	return f.Where(db.Or(filters...))
}

func (f *findResult) Populate(path string, options ...*db.PopulateOptions) db.FindResult {
	var opt *db.PopulateOptions
	if len(options) > 0 && options[0] != nil {
		opt = options[0]
	}
	f.populates[path] = opt
	return f
}

func (f *findResult) count(ignorePagination bool) (uint64, error) {
	countData := struct {
		Count uint64 `json:"_t"`
	}{}
	rows, err := f.executeQuery(sqlhelper.CountTemplate, func(data *findTemplateData) {
		if ignorePagination {
			data.Offset = 0
			data.Limit = 0
		}
	})
	if err != nil {
		f.err = err
		return countData.Count, err
	}
	if err := sqlhelper.One(rows, &countData); err != nil {
		f.err = err
		return countData.Count, err
	}
	return countData.Count, nil
}

func (f *findResult) Count() (uint64, error) {
	return f.count(false)
}

func (f *findResult) Delete() (uint64, error) {
	offset, limit := pagination(f.page, f.size)
	where, attrs := where(f.filters, f.meta)
	data := struct {
		Table  string
		Where  string
		Limit  uint
		Offset uint
	}{
		Table:  f.meta.NativeName,
		Where:  where,
		Limit:  limit,
		Offset: offset,
	}

	s, err := templatex.FastExecTXT(sqlhelper.DeleteTemplate, data)
	if err != nil {
		f.err = err
		return 0, err
	}

	result, err := f.common.Exec(s, attrs...)
	if err != nil {
		f.err = err
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		f.err = err
		return 0, err
	}
	return uint64(n), nil
}

func (f *findResult) Update(i interface{}) (uint64, error) {
	columns, values, err := sqlhelper.Map(i, nil)
	if err != nil {
		f.err = err
		return 0, err
	}

	var args []interface{}
	var vb bytes.Buffer
	for index, column := range columns {
		value := values[index]
		vb.WriteString(fmt.Sprintf("%s = ?", column))
		args = append(args, value)
		if index < len(columns)-1 {
			vb.WriteString(", ")
		}
	}
	where, attrs := where(f.filters, f.meta)
	columnValues := vb.String()
	data := struct {
		Table        string
		ColumnValues string
		Where        string
	}{
		Table:        f.meta.NativeName,
		ColumnValues: columnValues,
		Where:        where,
	}
	args = append(args, attrs...)

	s, err := templatex.FastExecTXT(sqlhelper.UpdateTemplate, data)
	if err != nil {
		f.err = err
		return 0, err
	}

	result, err := f.common.Exec(s, args...)
	if err != nil {
		f.err = err
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		f.err = err
		return 0, err
	}
	return uint64(n), nil
}

func (f *findResult) TotalPages() (uint, error) {
	var totalPages uint

	totalRecords, err := f.TotalRecords()
	if err != nil {
		return totalPages, err
	}

	totalPages = uint(math.Ceil(float64(totalRecords) / float64(f.size)))
	return totalPages, nil
}

func (f *findResult) TotalRecords() (uint64, error) {
	return f.count(true)
}
