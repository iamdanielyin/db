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
}

func (s *Scope) Skip() {
	s.skipLeft = true
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
