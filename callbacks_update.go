package db

func registerUpdateCallbacks(callbacks *clientWrapper) *clientWrapper {
	processor := callbacks.UpdateProcessors()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:before_update", beforeUpdateCallback)
	processor.Register("db:update", updateCallback)
	processor.Register("db:after_update", afterUpdateCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
	return callbacks
}

func beforeUpdateCallback(s *Scope) {
	s.callHooks(HookBeforeSave, s.Metadata.Name)
	s.callHooks(HookBeforeUpdate, s.Metadata.Name)
}

func updateCallback(s *Scope) {
	if s.HasError() {
		return
	}
	res := s.buildQueryResult()
	switch s.Action {
	case ActionUpdateOne:
		s.RecordsAffected, s.Error = res.UpdateOne(s.UpdateDoc)
	case ActionUpdateMany:
		s.RecordsAffected, s.Error = res.UpdateMany(s.UpdateDoc)
	}
}

func afterUpdateCallback(s *Scope) {
	s.callHooks(HookAfterUpdate, s.Metadata.Name)
	s.callHooks(HookAfterSave, s.Metadata.Name)
}
