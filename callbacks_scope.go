package db

import (
	"sync"
	"time"
)

type Scope struct {
	callbacks  *clientWrapper
	cacheStore *sync.Map
	skipLeft   bool

	Unscoped         bool
	Coll             Collection
	StartTime        time.Time
	Error            error
	Session          *Connection
	Metadata         Metadata
	Conditions       []Conditional
	PreloadsOptions  []*PreloadOptions
	Projection       []string
	Action           string
	OrderBys         []string
	PageSize         uint
	PageNum          uint
	InsertOneDoc     interface{}
	InsertManyDocs   interface{}
	UpdateDoc        interface{}
	UpdateOptions    *UpdateOptions
	InsertOneResult  InsertOneResult
	InsertManyResult InsertManyResult
	InsertOptions    *InsertOptions
	DeleteOptions    *DeleteOptions
	RecordsAffected  int
	TotalRecords     int
	TotalPages       int
	Dest             interface{}
	Cursor           Cursor
}

func (s *Scope) Skip() {
	s.skipLeft = true
}

func (s *Scope) AddCondition(v ...Conditional) *Scope {
	if len(v) == 0 {
		return s
	}
	for _, item := range v {
		if item != nil && len(item.Conditions()) > 0 {
			s.Conditions = append(s.Conditions)
		}
	}
	return s
}

func (s *Scope) And(i ...Conditional) *Scope {
	return s.AddCondition(And(i...))
}

func (s *Scope) Or(i ...Conditional) *Scope {
	return s.AddCondition(Or(i...))
}

func (s *Scope) HasError() bool {
	return s.Error != nil
}

func (s *Scope) Callbacks() *clientWrapper {
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
			s.AddCondition(rule.GetValue)
		}
	}
	var findArgs []interface{}
	if len(s.Conditions) > 0 {
		for _, item := range s.Conditions {
			findArgs = append(findArgs, item)
		}
	}
	res := s.Coll.Find(findArgs...)
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
			var ret bool
			switch s.Action {
			case ActionInsertOne:
				ret = testFieldsHook(hook, s.Action, s.InsertOneDoc)
			case ActionInsertMany:
				ret = testFieldsHook(hook, s.Action, s.InsertManyDocs)
			case ActionUpdateOne, ActionUpdateMany:
				ret = testFieldsHook(hook, s.Action, s.UpdateDoc)
			case ActionDeleteOne, ActionDeleteMany:
				ret = testFieldsHook(hook, s.Action, s.Conditions)
			}
			if ret {
				hook.Fn(s)
			}
		} else {
			hook.Fn(s)
		}
	}
}
