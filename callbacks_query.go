package db

func registerQueryCallbacks(callbacks *callbacks) *callbacks {
	processor := callbacks.Query()
	processor.Register("db:before_query", beforeQueryCallback)
	processor.Register("db:query", queryCallback)
	processor.Register("db:after_query", afterQueryCallback)
	return callbacks
}

func beforeQueryCallback(s *Scope) {
	s.callHooks(HookBeforeQuery, s.Metadata.Name)
}

func queryCallback(s *Scope) {
	if s.HasError() {
		return
	}

	res := s.buildQueryResult()

	switch s.Action {
	case ActionQueryOne:
		s.Error = res.One(s.Dest)
	case ActionQueryAll:
		s.Error = res.All(s.Dest)
	case ActionQueryCursor:
		s.Cursor, s.Error = res.Cursor()
	case ActionQueryCount:
		s.TotalRecords, s.Error = res.TotalRecords()
	case ActionQueryPage:
		s.TotalPages, s.Error = res.TotalPages()
	}
}

func afterQueryCallback(s *Scope) {
	s.callHooks(HookAfterQuery, s.Metadata.Name)
}
