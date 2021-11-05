package db

import (
	"sync"
	"time"
)

type Scope struct {
	callbacks  *callbacks
	cacheStore *sync.Map
	skipLeft   bool

	Unscoped         bool
	Coll             Collection
	StartTime        time.Time
	Error            error
	Session          *Connection
	Metadata         Metadata
	Conditions       []interface{}
	Projection       []string
	Action           string
	OrderBys         []string
	PageSize         uint
	PageNum          uint
	InsertOneDoc     interface{}
	InsertManyDocs   interface{}
	UpdateDoc        interface{}
	InsertOneResult  InsertOneResult
	InsertManyResult InsertManyResult
	RecordsAffected  int
	TotalRecords     int
	TotalPages       int
	Dest             interface{}
	Cursor           Cursor
}

func (s *Scope) Skip() {
	s.skipLeft = true
}

func (s *Scope) And(i ...interface{}) {
	s.Conditions = append(s.Conditions, And(i...))
}

func (s *Scope) Or(i ...interface{}) {
	s.Conditions = append(s.Conditions, Or(i...))
}

func (s *Scope) HasError() bool {
	return s.Error != nil
}

func (s *Scope) Callback() *callbacks {
	return s.callbacks
}

func (s *Scope) Store() *sync.Map {
	if s.cacheStore == nil {
		s.cacheStore = &sync.Map{}
	}
	return s.cacheStore
}

func (s *Scope) AddError(err error) *Scope {
	if err == nil {
		return s
	}
	if s.Error == nil {
		s.Error = err
	} else if err != nil {
		s.Error = Errorf("%v; %w", s.Error, err)
	}
	return s
}

func (s *Scope) buildQueryResult() Result {
	if !s.Unscoped {
		rule := LookupLogicDeleteRule(s.Metadata.Name)
		if rule != nil && rule.GetValue != nil {
			s.Conditions = append(s.Conditions, rule.GetValue)
		}
	}
	res := s.Coll.Find(s.Conditions...)
	if len(s.Projection) > 0 {
		res.Project(s.Projection...)
	}
	if len(s.OrderBys) > 0 {
		res.OrderBy(s.OrderBys...)
	}
	if s.PageSize > 0 {
		res.Paginate(s.PageSize)
	}
	if s.PageNum > 0 {
		res.Page(s.PageNum)
	}
	return res
}

func (s *Scope) callHooks(kind, name string) {
	if name == "" || kind == "" || s == nil {
		return
	}

	metadataHookMapMu.RLock()
	defer metadataHookMapMu.RUnlock()

	var hooks []*MetadataHook
	if v := metadataHookMap[name]; v != nil {
		hooks = v[kind]
	}

	for _, hook := range hooks {
		if len(hook.Fields) > 0 {
			var result bool
			switch s.Action {
			case ActionInsertOne:
				result = filterFields(hook, s.Action, s.InsertOneDoc)
			case ActionInsertMany:
				result = filterFields(hook, s.Action, s.InsertManyDocs)
			case ActionUpdateOne, ActionUpdateMany:
				result = filterFields(hook, s.Action, s.UpdateDoc)
			case ActionDeleteOne, ActionDeleteMany:
				result = filterFields(hook, s.Action, s.Conditions)
			}
			if !result {
				continue
			}
		}
		hook.Fn(s)
	}
}
