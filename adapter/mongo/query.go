package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"github.com/yuyitech/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"reflect"
	"strings"
	"time"
)

type mongoResult struct {
	mc         *mongoCollection
	conditions []interface{}
	projection []string
	orderBys   []string
	pageNum    uint
	pageSize   uint
	unscoped   bool
	filter     bson.D
}

func (r *mongoResult) And(i ...interface{}) db.Result {
	r.conditions = append(r.conditions, db.And(i...))
	return r
}

func (r *mongoResult) Or(i ...interface{}) db.Result {
	r.conditions = append(r.conditions, db.Or(i...))
	return r
}

func (r *mongoResult) Project(p ...string) db.Result {
	r.projection = p
	return r
}

func (r *mongoResult) One(dst interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := r.beforeQuery().mc.coll.FindOne(ctx,
		r.filter,
		r.buildFindOneOptions(),
	).Decode(dst)
	cancel()

	if err != nil && err != mongo.ErrNoDocuments {
		return db.Errorf(`%v`, err)
	}
	return nil
}

func (r *mongoResult) All(dst interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	cur, err := r.beforeQuery().mc.coll.Find(ctx,
		r.filter,
		r.buildFindOptions(),
	)
	cancel()
	if err != nil && err != mongo.ErrNoDocuments {
		return db.Errorf(`%v`, err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
	err = cur.All(ctx, dst)
	cancel()
	if err != nil {
		return db.Errorf(`%v`, err)
	}
	return nil
}

func (r *mongoResult) Cursor() (db.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	cur, err := r.beforeQuery().mc.coll.Find(ctx,
		r.filter,
		r.buildFindOptions(),
	)
	cancel()
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, db.Errorf(`%v`, err)
	}
	return &mongoCursor{result: r, cur: cur}, nil
}

func (r *mongoResult) OrderBy(s ...string) db.Result {
	r.orderBys = append(r.orderBys, s...)
	return r
}

func (r *mongoResult) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	val, err := r.beforeQuery().mc.coll.CountDocuments(ctx, r.filter)
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	return int(val), nil
}

func (r *mongoResult) Preload(i interface{}) db.Result {
	return r
}

func (r *mongoResult) Paginate(u uint) db.Result {
	r.pageSize = u
	return r
}

func (r *mongoResult) Page(u uint) db.Result {
	r.pageNum = u
	return r
}

func (r *mongoResult) TotalRecords() (int, error) {
	return r.Count()
}

func (r *mongoResult) TotalPages() (int, error) {
	if r.pageSize == 0 {
		return 1, nil
	}
	totalRecords, err := r.TotalRecords()
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(r.pageSize)))
	return totalPages, nil
}

func (r *mongoResult) Unscoped() db.Result {
	r.unscoped = true
	return r
}

func (r *mongoResult) UpdateOne(i interface{}, opts ...*db.UpdateOptions) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	doc := r.beforeUpdate(i)
	result, err := r.beforeQuery().mc.coll.UpdateOne(ctx,
		r.filter,
		doc,
	)
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	return int(result.MatchedCount), nil
}

func (r *mongoResult) UpdateMany(i interface{}, opts ...*db.UpdateOptions) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	doc := r.beforeUpdate(i)
	result, err := r.beforeQuery().mc.coll.UpdateMany(ctx,
		r.filter,
		doc,
	)
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	return int(result.MatchedCount), nil
}

func (r *mongoResult) DeleteOne(opts ...*db.DeleteOptions) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	result, err := r.beforeQuery().mc.coll.DeleteOne(ctx, r.filter)
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	return int(result.DeletedCount), nil
}

func (r *mongoResult) DeleteMany(opts ...*db.DeleteOptions) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	result, err := r.beforeQuery().mc.coll.DeleteMany(ctx,
		r.filter,
	)
	if err != nil {
		return 0, db.Errorf(`%v`, err)
	}
	return int(result.DeletedCount), nil
}

func (r *mongoResult) beforeQuery() *mongoResult {
	if r.filter == nil {
		r.filter = QueryFilter(r.mc.meta, r.conditions...)
	}
	if len(r.conditions) == 0 && r.filter == nil {
		r.filter = bson.D{}
	}
	return r
}

func (r *mongoResult) beforeUpdate(i interface{}) (result interface{}) {
	meta := r.mc.meta
	reflectValue := reflect.Indirect(reflect.ValueOf(i))
	switch reflectValue.Kind() {
	case reflect.Struct:
		s := structs.New(i)
		var doc bson.D
		for _, field := range s.Fields() {
			if field.IsZero() {
				continue
			}
			key := field.Name()
			if f, has := meta.FieldByName(key); has {
				key = f.MustNativeName()
			}
			doc = append(doc, bson.E{Key: key, Value: field.Value()})
		}
		result = bson.D{bson.E{Key: "$set", Value: doc}}
	case reflect.Map:
		var doc bson.D
		for _, k := range reflectValue.MapKeys() {
			key := k.Interface().(string)
			val := reflectValue.MapIndex(k).Interface()
			if f, has := meta.FieldByName(key); has {
				key = f.MustNativeName()
			}
			doc = append(doc, bson.E{Key: key, Value: val})
		}
		result = bson.D{bson.E{Key: "$set", Value: doc}}
	default:
		result = i
	}
	return
}

func (r *mongoResult) buildFindOptions() *options.FindOptions {
	opts := options.Find()
	if r.pageSize > 0 {
		opts.SetLimit(int64(r.pageSize))
		if r.pageNum > 0 {
			opts.SetSkip(int64((r.pageNum - 1) * r.pageSize))
		}
	}
	meta := r.mc.meta
	if len(r.orderBys) > 0 {
		var sort bson.D
		for _, item := range r.orderBys {
			var (
				key   = item
				value = 1
			)
			if strings.HasPrefix(item, "-") {
				key = item[1:]
				value = -1
			}
			if f, has := meta.FieldByName(key); has {
				key = f.MustNativeName()
			}
			sort = append(sort, bson.E{Key: key, Value: value})
		}
		if len(sort) > 0 {
			opts.SetSort(sort)
		}
	}
	if len(r.projection) > 0 {
		var projection bson.D
		for _, item := range r.projection {
			var (
				key   = item
				value = 1
			)
			if strings.HasPrefix(item, "-") {
				key = item[1:]
				value = 0
			}
			if f, has := meta.FieldByName(key); has {
				key = f.MustNativeName()
			}
			projection = append(projection, bson.E{Key: key, Value: value})
		}
		if len(projection) > 0 {
			opts.SetProjection(projection)
		}
	}
	return opts
}

func (r *mongoResult) buildFindOneOptions() *options.FindOneOptions {
	findOpts := r.buildFindOptions()
	return &options.FindOneOptions{
		Projection: findOpts.Projection,
		Skip:       findOpts.Skip,
		Sort:       findOpts.Sort,
	}
}

type mongoCursor struct {
	result          *mongoResult
	cur             *mongo.Cursor
	unprocessedNext bool
	lastNextValue   bool
}

func (c *mongoCursor) HasNext() bool {
	if c.unprocessedNext {
		return c.lastNextValue
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	c.unprocessedNext = true
	c.lastNextValue = c.cur.Next(ctx)
	return c.lastNextValue
}

func (c *mongoCursor) Next(dst interface{}) error {
	c.unprocessedNext = true
	if err := c.cur.Decode(dst); err != nil {
		return db.Errorf(`%v`, err)
	}
	return nil
}

func (c *mongoCursor) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	if err := c.cur.Close(ctx); err != nil {
		return db.Errorf(`%v`, err)
	}
	return nil
}
