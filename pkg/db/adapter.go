package db

import "github.com/yuyitech/db/pkg/schema"

type database struct {
	target     Database
	middleware *AdapterMiddleware
}

func (d *database) Model(s string) Collection {
	if d.target == nil {
		d.target = Session()
	}
	m := d.target.Model(s)
	return &model{target: m}
}

func (d *database) Name() string {
	return d.target.Name()
}

func (d *database) Query(s string, i ...interface{}) Query {
	return d.target.Query(s, i...)
}

func (d *database) Exec(s string, i ...interface{}) (interface{}, uint64, error) {
	return d.target.Exec(s, i...)
}

func (d *database) DataSource() *DataSource {
	return d.target.DataSource()
}

func (d *database) DriverName() string {
	return d.target.DriverName()
}

func (d *database) Open(source *DataSource) (Database, error) {
	return d.target.Open(source)
}

func (d *database) Close() error {
	return d.target.Close()
}

func (d *database) NativeCollectionNames() ([]string, error) {
	return d.target.NativeCollectionNames()
}

func (d *database) NativeCollectionMetadata() ([]schema.Metadata, error) {
	return d.target.NativeCollectionMetadata()
}

func (d *database) BeginTx() (Tx, error) {
	t, err := d.target.BeginTx()
	if err != nil {
		return nil, err
	}
	return &tx{target: t}, nil
}

type model struct {
	target Collection
}

func (m *model) Name() string {
	return m.target.Name()
}

func (m *model) Metadata() schema.Metadata {
	return m.target.Metadata()
}

func (m *model) Database() Database {
	return m.target.Database()
}
func (m *model) Middleware() *AdapterMiddleware {
	if v := m.target.Middleware(); v != nil {
		return v
	}
	return DefaultAdapterMiddleware
}

func (m *model) Create(i interface{}) (interface{}, uint64, error) {
	scope := m.Middleware().NewScope(m.target, i).callMiddleware(m.Middleware().creates)
	return scope.OutputValue, scope.RecordsAffected, scope.Error
}

func (m *model) Find(i ...interface{}) FindResult {
	fr := m.target.Find(i...)
	search := &findResult{target: fr, middleware: m.Middleware()}
	scope := search.middleware.NewScope(m.target, nil)
	scope.search = fr
	search.scope = scope
	return search
}

type tx struct {
	target Tx
}

func (t *tx) Model(s string) Collection {
	m := t.target.Model(s)
	if m != nil {
		return &model{target: m}
	}
	return nil
}

func (t *tx) Name() string {
	return t.target.Name()
}

func (t *tx) Query(s string, i ...interface{}) Query {
	return t.target.Query(s, i...)
}
func (t *tx) Exec(s string, i ...interface{}) (interface{}, uint64, error) {
	return t.target.Exec(s, i...)
}

func (t *tx) DataSource() *DataSource {
	return t.target.DataSource()
}

func (t *tx) Rollback() error {
	return t.target.Rollback()
}

func (t *tx) Commit() error {
	return t.target.Commit()
}

type findResult struct {
	scope      *AdapterMiddlewareScope
	middleware *AdapterMiddleware
	target     FindResult
}

func (f *findResult) Page(u uint) FindResult {
	return f.target.Page(u)
}

func (f *findResult) Size(u uint) FindResult {
	return f.target.Size(u)
}

func (f *findResult) Order(s ...string) FindResult {
	return f.target.Order(s...)
}

func (f *findResult) Select(s ...string) FindResult {
	return f.target.Select(s...)
}

func (f *findResult) Where(i interface{}) FindResult {
	return f.target.Where(i)
}

func (f *findResult) And(cond ...Cond) FindResult {
	return f.target.And(cond...)
}

func (f *findResult) Or(cond ...Cond) FindResult {
	return f.target.Or(cond...)
}

func (f *findResult) Iterator() (Iterator, error) {
	f.scope.queryAction = queryActionIterator
	scope := f.scope.callMiddleware(f.middleware.queries)
	return scope.Iterator, scope.Error
}

func (f *findResult) One(ptrToStruct interface{}) error {
	f.scope.dst = ptrToStruct
	f.scope.queryAction = queryActionOne
	scope := f.scope.callMiddleware(f.middleware.queries)
	return scope.Error
}

func (f *findResult) All(sliceOfStruct interface{}) error {
	f.scope.dst = sliceOfStruct
	f.scope.queryAction = queryActionAll
	scope := f.scope.callMiddleware(f.middleware.queries)
	return scope.Error
}

func (f *findResult) Populate(s string, options ...*PopulateOptions) FindResult {
	return f.target.Populate(s, options...)
}

func (f *findResult) Count() (uint64, error) {
	f.scope.queryAction = queryActionCount
	scope := f.scope.callMiddleware(f.middleware.queries)
	return scope.RecordsAffected, scope.Error
}

func (f *findResult) Delete() (uint64, error) {
	scope := f.scope.callMiddleware(f.middleware.deletes)
	return scope.RecordsAffected, scope.Error
}

func (f *findResult) Update(i interface{}) (uint64, error) {
	f.scope.InputValue = i
	scope := f.scope.callMiddleware(f.middleware.updates)
	return scope.RecordsAffected, scope.Error
}

func (f *findResult) TotalPages() (uint, error) {
	return f.target.TotalPages()
}

func (f *findResult) TotalRecords() (uint64, error) {
	return f.target.TotalRecords()
}
