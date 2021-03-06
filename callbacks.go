package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	ActionInsertOne   = "INSERT_ONE"
	ActionInsertMany  = "INSERT_MANY"
	ActionUpdateOne   = "UPDATE_ONE"
	ActionUpdateMany  = "UPDATE_MANY"
	ActionDeleteOne   = "DELETE_ONE"
	ActionDeleteMany  = "DELETE_MANY"
	ActionQueryOne    = "QUERY_ONE"
	ActionQueryAll    = "QUERY_ALL"
	ActionQueryCursor = "QUERY_CURSOR"
	ActionQueryCount  = "QUERY_COUNT"
	ActionQueryPage   = "QUERY_PAGE"
)

func newClientWrapper(raw Client, sess *Connection) *clientWrapper {
	return &clientWrapper{
		rawClient: raw,
		processors: map[string]*processor{
			"create": {sess: sess},
			"query":  {sess: sess},
			"update": {sess: sess},
			"delete": {sess: sess},
			"row":    {sess: sess},
			"raw":    {sess: sess},
		},
	}
}

type clientWrapper struct {
	processors map[string]*processor
	rawClient  Client
}

func (cs *clientWrapper) Name() string {
	return cs.rawClient.Name()
}

func (cs *clientWrapper) Logger() Logger {
	return cs.rawClient.Logger()
}

func (cs *clientWrapper) Source() DataSource {
	return cs.rawClient.Source()
}

func (cs *clientWrapper) Raw(raw string, values ...interface{}) error {
	return cs.rawClient.Raw(raw, values...)
}

func (cs *clientWrapper) Disconnect(ctx context.Context) error {
	return cs.rawClient.Disconnect(ctx)
}

func (cs *clientWrapper) Model(metadata Metadata) Collection {
	rawColl := cs.rawClient.Model(metadata)
	return &callbacksCollection{client: cs, rawColl: rawColl}
}

func (cs *clientWrapper) StartTransaction() (Tx, error) {
	return cs.rawClient.StartTransaction()
}

func (cs *clientWrapper) WithTransaction(f func(Tx) error) error {
	return cs.rawClient.WithTransaction(f)
}

type callbacksCollection struct {
	client  *clientWrapper
	rawColl Collection
}

func (cc *callbacksCollection) NewScope(scope *Scope) *Scope {
	scope.StartTime = time.Now()
	scope.Session = cc.Session()
	scope.Metadata = cc.Metadata()
	if scope.cacheStore == nil {
		scope.cacheStore = &sync.Map{}
	}
	return scope
}

func (cc *callbacksCollection) Name() string {
	return cc.rawColl.Name()
}

func (cc *callbacksCollection) Metadata() Metadata {
	return cc.rawColl.Metadata()
}

func (cc *callbacksCollection) Session() *Connection {
	return cc.rawColl.Session()
}

func (cc *callbacksCollection) InsertOne(i interface{}, fns ...func(*InsertOptions)) (InsertOneResult, error) {
	scope := &Scope{
		Action:       ActionInsertOne,
		InsertOneDoc: i,
		Coll:         cc.rawColl,
	}
	if len(fns) > 0 && fns[0] != nil {
		scope.InsertOptions = new(InsertOptions)
		for _, fn := range fns {
			fn(scope.InsertOptions)
		}
	}
	cc.client.CreateProcessors().Execute(cc.NewScope(scope))
	return scope.InsertOneResult, scope.Error
}

func (cc *callbacksCollection) InsertMany(i interface{}, fns ...func(*InsertOptions)) (InsertManyResult, error) {
	scope := &Scope{
		Action:         ActionInsertMany,
		InsertManyDocs: i,
		Coll:           cc.rawColl,
	}
	if len(fns) > 0 && fns[0] != nil {
		scope.InsertOptions = new(InsertOptions)
		for _, fn := range fns {
			fn(scope.InsertOptions)
		}
	}
	cc.client.CreateProcessors().Execute(cc.NewScope(scope))
	return scope.InsertManyResult, scope.Error
}

func (cc *callbacksCollection) testKeyValuePairs(v []interface{}) bool {
	if len(v) > 0 && len(v)%2 == 0 {
		if _, ok := v[0].(string); !ok {
			return true
		}
	}
	return false
}

func (cc *callbacksCollection) parseKeyValuePairs(v []interface{}) *Cond {
	var (
		max  = len(v)
		cond = make(Cond)
	)
	for i := 0; i < max; i++ {
		if i%2 != 0 {
			continue
		}
		if key, ok := v[i].(string); ok {
			var value interface{}
			if i+1 < max {
				value = v[i+1]
			}
			if !IsNil(value) {
				cond[key] = value
			}
		}
	}

	if len(cond) > 0 {
		return &cond
	}
	return nil
}

func (cc *callbacksCollection) Find(v ...interface{}) Result {
	var conditions []Conditional
	if cc.testKeyValuePairs(v) {
		if cond := cc.parseKeyValuePairs(v); cond != nil {
			conditions = append(conditions, cond)
		}
	} else {
		for _, item := range v {
			if vv, ok := item.(Conditional); ok {
				if vv != nil && len(vv.Conditions()) > 0 {
					conditions = append(conditions, vv)
				}
			}
		}
	}
	scope := &Scope{
		Coll:       cc.rawColl,
		Conditions: conditions,
	}
	return &callbacksResult{
		cc:    cc,
		scope: scope,
	}
}

type callbacksResult struct {
	cc    *callbacksCollection
	scope *Scope
}

func (cr *callbacksResult) And(i ...Conditional) Result {
	cr.scope.AddCondition(And(i...))
	return cr
}

func (cr *callbacksResult) Or(i ...Conditional) Result {
	cr.scope.Conditions = append(cr.scope.Conditions, Or(i...))
	return cr
}

func (cr *callbacksResult) Project(p ...string) Result {
	cr.scope.Projection = p
	return cr
}

func (cr *callbacksResult) OrderBy(s ...string) Result {
	cr.scope.OrderBys = append(cr.scope.OrderBys, s...)
	return cr
}

func (cr *callbacksResult) Paginate(u uint) Result {
	cr.scope.PageSize = u
	return cr
}

func (cr *callbacksResult) Page(u uint) Result {
	cr.scope.PageNum = u
	return cr
}

func (cr *callbacksResult) Unscoped() Result {
	cr.scope.Unscoped = true
	return cr
}

func (cr *callbacksResult) Preload(path string, fns ...func(options *PreloadOptions)) Result {
	if path != "" && len(fns) > 0 {
		if cr.scope.Preloads == nil {
			cr.scope.Preloads = make(map[string][]func(options *PreloadOptions))
		}
		cr.scope.Preloads[path] = fns
	}
	return cr
}

func (cr *callbacksResult) One(dst interface{}) error {
	cr.scope.Dest = dst
	cr.scope.Action = ActionQueryOne
	cr.cc.client.QueryProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.Error
}

func (cr *callbacksResult) All(dst interface{}) error {
	cr.scope.Dest = dst
	cr.scope.Action = ActionQueryAll
	cr.cc.client.QueryProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.Error
}

func (cr *callbacksResult) Cursor() (Cursor, error) {
	cr.scope.Action = ActionQueryCursor
	cr.cc.client.QueryProcessors().Execute(cr.cc.NewScope(cr.scope))
	return &callbacksCursor{cr: cr, rawCursor: cr.scope.Cursor}, cr.scope.Error
}

func (cr *callbacksResult) Count() (int, error) {
	cr.scope.Action = ActionQueryCount
	cr.cc.client.QueryProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.TotalRecords, cr.scope.Error
}

func (cr *callbacksResult) TotalRecords() (int, error) {
	return cr.Count()
}

func (cr *callbacksResult) TotalPages() (int, error) {
	cr.scope.Action = ActionQueryPage
	cr.cc.client.QueryProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.TotalPages, cr.scope.Error
}

func (cr *callbacksResult) UpdateOne(i interface{}, fns ...func(*UpdateOptions)) (int, error) {
	cr.scope.Action = ActionUpdateOne
	cr.scope.UpdateDoc = i
	if len(fns) > 0 && fns[0] != nil {
		cr.scope.UpdateOptions = new(UpdateOptions)
		for _, fn := range fns {
			fn(cr.scope.UpdateOptions)
		}
	}
	cr.cc.client.UpdateProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.RecordsAffected, cr.scope.Error
}

func (cr *callbacksResult) UpdateMany(i interface{}, fns ...func(*UpdateOptions)) (int, error) {
	cr.scope.Action = ActionUpdateMany
	cr.scope.UpdateDoc = i
	if len(fns) > 0 && fns[0] != nil {
		cr.scope.UpdateOptions = new(UpdateOptions)
		for _, fn := range fns {
			fn(cr.scope.UpdateOptions)
		}
	}
	cr.cc.client.UpdateProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.RecordsAffected, cr.scope.Error
}

func (cr *callbacksResult) DeleteOne(fns ...func(*DeleteOptions)) (int, error) {
	cr.scope.Action = ActionDeleteOne
	if len(fns) > 0 && fns[0] != nil {
		cr.scope.DeleteOptions = new(DeleteOptions)
		for _, fn := range fns {
			fn(cr.scope.DeleteOptions)
		}
	}
	cr.cc.client.DeleteProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.RecordsAffected, cr.scope.Error
}

func (cr *callbacksResult) DeleteMany(fns ...func(*DeleteOptions)) (int, error) {
	cr.scope.Action = ActionDeleteMany
	if len(fns) > 0 && fns[0] != nil {
		cr.scope.DeleteOptions = new(DeleteOptions)
		for _, fn := range fns {
			fn(cr.scope.DeleteOptions)
		}
	}
	cr.cc.client.DeleteProcessors().Execute(cr.cc.NewScope(cr.scope))
	return cr.scope.RecordsAffected, cr.scope.Error
}

type callbacksCursor struct {
	cr        *callbacksResult
	rawCursor Cursor
}

func (c *callbacksCursor) HasNext() bool {
	return c.rawCursor.HasNext()
}

func (c *callbacksCursor) Next(dst interface{}) error {
	return c.rawCursor.Next(dst)
}

func (c *callbacksCursor) Close() error {
	return c.rawCursor.Close()
}

type processor struct {
	sess      *Connection
	fns       []func(*Scope)
	callbacks []*callback
}

type callback struct {
	name      string
	before    string
	after     string
	remove    bool
	replace   bool
	match     func(*Scope) bool
	handler   func(*Scope)
	processor *processor
}

func (cs *clientWrapper) CreateProcessors() *processor {
	return cs.processors["create"]
}

func (cs *clientWrapper) QueryProcessors() *processor {
	return cs.processors["query"]
}

func (cs *clientWrapper) UpdateProcessors() *processor {
	return cs.processors["update"]
}

func (cs *clientWrapper) DeleteProcessors() *processor {
	return cs.processors["delete"]
}

func (cs *clientWrapper) RowProcessors() *processor {
	return cs.processors["row"]
}

func (cs *clientWrapper) RawProcessors() *processor {
	return cs.processors["raw"]
}

func (p *processor) Execute(s *Scope) {
	for _, f := range p.fns {
		f(s)
		if s.skipLeft {
			break
		}
	}
}

func (p *processor) Get(name string) func(*Scope) {
	for i := len(p.callbacks) - 1; i >= 0; i-- {
		if v := p.callbacks[i]; v.name == name && !v.remove {
			return v.handler
		}
	}
	return nil
}

func (p *processor) Before(name string) *callback {
	return &callback{before: name, processor: p}
}

func (p *processor) After(name string) *callback {
	return &callback{after: name, processor: p}
}

func (p *processor) Match(fc func(*Scope) bool) *callback {
	return &callback{match: fc, processor: p}
}

func (p *processor) Register(name string, fn func(*Scope)) {
	(&callback{processor: p}).Register(name, fn)
}

func (p *processor) Remove(name string) {
	(&callback{processor: p}).Remove(name)
}

func (p *processor) Replace(name string, fn func(*Scope)) {
	(&callback{processor: p}).Replace(name, fn)
}

func (p *processor) compile() {
	var callbacks []*callback
	s := &Scope{Session: p.sess}
	for _, callback := range p.callbacks {
		if callback.match == nil || callback.match(s) {
			callbacks = append(callbacks, callback)
		}
	}
	p.callbacks = callbacks

	var err error
	if p.fns, err = sortCallbacks(p.callbacks); err != nil {
		panic(Errorf("compile callbacks error %v", err))
	}
	return
}

func (c *callback) Before(name string) *callback {
	c.before = name
	return c
}

func (c *callback) After(name string) *callback {
	c.after = name
	return c
}

func (c *callback) Register(name string, fn func(*Scope)) {
	c.name = name
	c.handler = fn
	c.processor.callbacks = append(c.processor.callbacks, c)
	c.processor.compile()
}

func (c *callback) Remove(name string) {
	c.name = name
	c.remove = true
	c.processor.callbacks = append(c.processor.callbacks, c)
	c.processor.compile()
}

func (c *callback) Replace(name string, fn func(*Scope)) {
	c.name = name
	c.handler = fn
	c.replace = true
	c.processor.callbacks = append(c.processor.callbacks, c)
	c.processor.compile()
}

// getRIndex get right index from string slice
func getRIndex(strs []string, str string) int {
	for i := len(strs) - 1; i >= 0; i-- {
		if strs[i] == str {
			return i
		}
	}
	return -1
}

func sortCallbacks(cs []*callback) (fns []func(*Scope), err error) {
	var (
		names, sorted []string
		sortCallback  func(*callback) error
	)
	sort.Slice(cs, func(i, j int) bool {
		return cs[j].before == "*" || cs[j].after == "*"
	})

	for _, c := range cs {
		names = append(names, c.name)
	}

	sortCallback = func(c *callback) error {
		if c.before != "" { // if defined before callback
			if c.before == "*" && len(sorted) > 0 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					sorted = append([]string{c.name}, sorted...)
				}
			} else if sortedIdx := getRIndex(sorted, c.before); sortedIdx != -1 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					// if before callback already sorted, append current callback just after it
					sorted = append(sorted[:sortedIdx], append([]string{c.name}, sorted[sortedIdx:]...)...)
				} else if curIdx > sortedIdx {
					return fmt.Errorf("conflicting callback %s with before %s", c.name, c.before)
				}
			} else if idx := getRIndex(names, c.before); idx != -1 {
				// if before callback exists
				cs[idx].after = c.name
			}
		}

		if c.after != "" { // if defined after callback
			if c.after == "*" && len(sorted) > 0 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					sorted = append(sorted, c.name)
				}
			} else if sortedIdx := getRIndex(sorted, c.after); sortedIdx != -1 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					// if after callback sorted, append current callback to last
					sorted = append(sorted, c.name)
				} else if curIdx < sortedIdx {
					return fmt.Errorf("conflicting callback %s with before %s", c.name, c.after)
				}
			} else if idx := getRIndex(names, c.after); idx != -1 {
				// if after callback exists but haven't sorted
				// set after callback's before callback to current callback
				after := cs[idx]

				if after.before == "" {
					after.before = c.name
				}

				if err := sortCallback(after); err != nil {
					return err
				}

				if err := sortCallback(c); err != nil {
					return err
				}
			}
		}

		// if current callback haven't been sorted, append it to last
		if getRIndex(sorted, c.name) == -1 {
			sorted = append(sorted, c.name)
		}

		return nil
	}

	for _, c := range cs {
		if err = sortCallback(c); err != nil {
			return
		}
	}

	for _, name := range sorted {
		if idx := getRIndex(names, name); !cs[idx].remove {
			fns = append(fns, cs[idx].handler)
		}
	}

	return
}
