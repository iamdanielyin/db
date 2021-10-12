package mongo

import "github.com/yuyitech/db"

type mongoResult struct {
	coll       *mongoCollection
	conditions []interface{}
	project    map[string]int
	orderBys   []string
	pageNum    uint
	pageSize   uint
	unscoped   bool
}

func (r *mongoResult) And(i ...interface{}) db.Result {
	r.conditions = append(r.conditions, db.And(i...))
	return r
}

func (r *mongoResult) Or(i ...interface{}) db.Result {
	r.conditions = append(r.conditions, db.Or(i...))
	return r
}

func (r *mongoResult) Project(m map[string]int) db.Result {
	r.project = m
	return r
}

func (r *mongoResult) One(dst interface{}) error {
	panic("implement me")
}

func (r *mongoResult) All(dst interface{}) error {
	panic("implement me")
}

func (r *mongoResult) Cursor() db.Cursor {
	panic("implement me")
}

func (r *mongoResult) OrderBy(s ...string) db.Result {
	r.orderBys = append(r.orderBys, s...)
	return r
}

func (r *mongoResult) Count() (int, error) {
	panic("implement me")
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
	panic("implement me")
}

func (r *mongoResult) TotalPages() (int, error) {
	panic("implement me")
}

func (r *mongoResult) UpdateOne(i interface{}) (db.UpdateResult, error) {
	panic("implement me")
}

func (r *mongoResult) UpdateMany(i interface{}) (db.UpdateResult, error) {
	panic("implement me")
}

func (r *mongoResult) Unscoped() db.Result {
	r.unscoped = true
	return r
}

func (r *mongoResult) DeleteOne(i interface{}) (db.DeleteResult, error) {
	panic("implement me")
}

func (r *mongoResult) DeleteMany(i interface{}) (db.DeleteResult, error) {
	panic("implement me")
}
