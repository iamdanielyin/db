package db

func registerDeleteCallbacks(callbacks *callbacks) *callbacks {
	processor := callbacks.Delete()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:delete", deleteCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
	return callbacks
}

func deleteCallback(s *Scope) {
	if s.HasError() {
		return
	}
	res := buildQueryResult(s)
	switch s.Action {
	case ActionDeleteOne:
		s.RecordsAffected, s.Error = res.DeleteOne()
	case ActionDeleteMany:
		s.RecordsAffected, s.Error = res.DeleteMany()
	}
}
