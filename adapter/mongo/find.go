package mongo

import (
	"context"
	"fmt"
	"github.com/yuyitech/db/internal/json"
	"github.com/yuyitech/db/internal/reflectx"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"reflect"
	"strings"
	"time"
)

type findResult struct {
	m    *model
	a    *adapter
	coll *mongo.Collection
	meta db.Metadata

	cur *mongo.Cursor
	ctx context.Context

	err       error
	page      uint
	size      uint
	filters   []interface{}
	orders    map[string]bool
	selects   map[string]bool
	populates map[string]*db.PopulateOptions
}

func newFindResult(m *model, filter []interface{}) *findResult {
	result := &findResult{
		m:    m,
		a:    m.a,
		coll: m.coll,
		meta: m.meta,

		orders:    make(map[string]bool),
		selects:   make(map[string]bool),
		populates: make(map[string]*db.PopulateOptions),
	}
	if len(filter) > 0 {
		for _, item := range filter {
			if item != nil {
				result.filters = append(result.filters, item)
			}
		}
	}
	return result
}

func (f *findResult) Iterator() (db.Iterator, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	cur, err := f.buildCursor(f.ctx)
	if err != nil {
		return nil, err
	}
	return &iterator{
		ctx: ctx,
		err: err,
		cur: cur,
	}, err
}

func where(filters []interface{}, meta db.Metadata) *bson.D {
	filter := ParseFilter(func(cmp *db.Comparison) {
		field := meta.Fields[cmp.Key]
		cmp.Key = field.MustNativeName()
	}, filters...)
	logger.INFO(json.Stringify(filter, false))
	return filter
}

func pagination(page, size uint) (offset int64, limit int64) {
	if size > 0 && page > 0 {
		offset = int64(page - 1)
		limit = int64(size)
	}
	return
}

func projection(selects map[string]bool, meta db.Metadata) bson.M {
	project := make(bson.M)
	for k, v := range selects {
		field := meta.Fields[k]
		nativeName := field.MustNativeName()
		if v {
			project[nativeName] = 1
		} else {
			project[nativeName] = 0
		}
	}
	return project
}

func sort(orders map[string]bool, meta db.Metadata) bson.M {
	sort := make(bson.M)
	for k, v := range orders {
		field := meta.Fields[k]
		nativeName := field.MustNativeName()
		if v {
			sort[nativeName] = -1
		} else {
			sort[nativeName] = 1
		}
	}
	return sort
}
func (f *findResult) populateOne(ptrToStruct interface{}) {
	dstReflectValue := reflect.ValueOf(ptrToStruct)
	mapper := reflectx.NewMapper("json")

	if len(f.populates) > 0 {
		for path, opts := range f.populates {
			field := f.meta.Fields[path]
			if field.Name == "" {
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
	filter := where(f.filters, f.meta)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	f.err = f.coll.FindOne(ctx, *filter).Decode(ptrToStruct)
	f.populateOne(ptrToStruct)
	return f.err
}

func (f *findResult) buildCursor(ctx context.Context, cb ...func(*options.FindOptions)) (*mongo.Cursor, error) {
	filter := where(f.filters, f.meta)
	offset, limit := pagination(f.page, f.size)
	project := projection(f.selects, f.meta)
	sort := sort(f.orders, f.meta)
	opts := &options.FindOptions{
		Skip:       &offset,
		Limit:      &limit,
		Projection: project,
		Sort:       sort,
	}
	if len(cb) > 0 && cb[0] != nil {
		cb[0](opts)
	}
	return f.coll.Find(ctx, *filter, opts)
}

func (f *findResult) populateAll(sliceOfStruct interface{}) {
	mapper := reflectx.NewMapper("json")
	dstReflectValue := reflect.ValueOf(sliceOfStruct)
	dstIndirectValue := reflect.Indirect(dstReflectValue)

	if len(f.populates) > 0 && dstIndirectValue.Len() > 0 {
		for path, opts := range f.populates {
			populateField := f.meta.Fields[path]
			if populateField.Name == "" {
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
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	cur, err := f.buildCursor(ctx)
	if err != nil {
		f.err = err
		return f.err
	}
	f.err = cur.All(ctx, sliceOfStruct)
	f.populateAll(sliceOfStruct)
	return f.err
}

func (f *findResult) Page(u uint) db.IFindResult {
	f.page = u
	return f
}

func (f *findResult) Size(u uint) db.IFindResult {
	f.size = u
	return f
}

func (f *findResult) Order(s ...string) db.IFindResult {
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

func (f *findResult) Select(s ...string) db.IFindResult {
	for _, item := range s {
		sel := true
		if strings.HasPrefix(item, "-") {
			sel = false
			item = item[1:]
		}
		f.selects[item] = sel
	}
	return f
}

func (f *findResult) Where(filter interface{}) db.IFindResult {
	f.filters = append(f.filters, filter)
	return f
}

func (f *findResult) And(filters ...db.Cond) db.IFindResult {
	return f.Where(db.And(filters...))
}

func (f *findResult) Or(filters ...db.Cond) db.IFindResult {
	return f.Where(db.Or(filters...))
}

func (f *findResult) Populate(path string, options ...*db.PopulateOptions) db.IFindResult {
	var opt *db.PopulateOptions
	if len(options) > 0 && options[0] != nil {
		opt = options[0]
	}
	f.populates[path] = opt
	return f
}

func (f *findResult) count(ignorePagination bool) (uint64, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	filter := where(f.filters, f.meta)

	var (
		offset int64
		limit  int64
	)
	if !ignorePagination {
		offset, limit = pagination(f.page, f.size)
	}
	opts := &options.CountOptions{
		Skip:  &offset,
		Limit: &limit,
	}

	count, err := f.coll.CountDocuments(ctx, *filter, opts)
	if err != nil {
		f.err = err
	}
	return uint64(count), f.err
}

func (f *findResult) Count() (uint64, error) {
	return f.count(false)
}

func (f *findResult) Delete() (uint64, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	filter := where(f.filters, f.meta)

	res, err := f.coll.DeleteMany(ctx, filter)
	if err != nil {
		f.err = err
		return 0, f.err
	}

	return uint64(res.DeletedCount), f.err
}

func (f *findResult) Update(i interface{}) (uint64, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	filter := where(f.filters, f.meta)

	res, err := f.coll.UpdateMany(ctx, filter, i)
	if err != nil {
		f.err = err
		return 0, f.err
	}

	return uint64(res.ModifiedCount), f.err
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
