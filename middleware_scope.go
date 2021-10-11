package db

import "time"

type Scope struct {
	StartTime        time.Time
	Error            error
	IsSkip           bool
	Session          *Connection
	Metadata         Metadata
	Conditions       []Cond
	Action           string
	OrderBys         []string
	PageSize         int
	PageNum          int
	InsertOneDoc     interface{}
	InsertManyDocs   interface{}
	UpdateDoc        interface{}
	InsertOneResult  InsertOneResult
	InsertManyResult InsertManyResult
	UpdateOneResult  UpdateOneResult
	UpdateManyResult UpdateManyResult
	DeleteOneResult  DeleteOneResult
	DeleteManyResult DeleteManyResult
}

func (s *Scope) Skip() {
	s.IsSkip = true
}
func (s *Scope) HasError() {
	s.IsSkip = true
}

func (s *Scope) AddError(err error) error {
	if s.Error == nil {
		s.Error = err
	} else if err != nil {
		s.Error = Errorf("%v; %w", s.Error, err)
	}
	return s.Error
}
