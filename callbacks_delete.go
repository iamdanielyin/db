package db

func registerDeleteCallbacks(callbacks *callbacks) *callbacks {
	processor := callbacks.Delete()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:before_delete", beforeDeleteCallback)
	processor.Register("db:delete", deleteCallback)
	processor.Register("db:after_delete", afterDeleteCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
	return callbacks
}

func beforeDeleteCallback(s *Scope) {
	s.callHooks(HookBeforeDelete, s.Metadata.Name)
}

func deleteCallback(s *Scope) {
	if s.HasError() {
		return
	}
	res := s.buildQueryResult()
	switch s.Action {
	case ActionDeleteOne:
		s.RecordsAffected, s.Error = res.DeleteOne()
	case ActionDeleteMany:
		s.RecordsAffected, s.Error = res.DeleteMany()
	}
}

func afterDeleteCallback(s *Scope) {
	s.callHooks(HookAfterDelete, s.Metadata.Name)
}
