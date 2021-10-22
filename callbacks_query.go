package db

func registerQueryCallbacks(callbacks *callbacks) *callbacks {
	return callbacks
}

func buildQueryResult(s *Scope) Result {
	res := s.Coll.Find(s.Conditions...)
	if s.Unscoped {
		res.Unscoped()
	}
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
