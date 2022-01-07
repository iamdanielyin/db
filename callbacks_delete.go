package db

func registerDeleteCallbacks(callbacks *clientWrapper) *clientWrapper {
	processor := callbacks.DeleteProcessors()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:before_delete", beforeDeleteCallback)
	processor.Register("db:logic_delete", logicDeleteCallback)
	processor.Register("db:delete", deleteCallback)
	processor.Register("db:after_delete", afterDeleteCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
	return callbacks
}

func logicDeleteCallback(s *Scope) {
	if s.Unscoped {
		return
	}

	rule := LookupLogicDeleteRule(s.Metadata.Name)
	if rule == nil {
		return
	}

	if values := rule.ParseSetValue(); values != nil {
		var doc = make(map[string]interface{})
		for key, val := range values {
			key = s.Metadata.MustFieldNativeName(key)
			doc[key] = val
		}
		s.UpdateDoc = doc
	}
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
		if s.UpdateDoc != nil {
			s.RecordsAffected, s.Error = res.UpdateOne(s.UpdateDoc)
		} else {
			s.RecordsAffected, s.Error = res.DeleteOne()
		}
	case ActionDeleteMany:
		if s.UpdateDoc != nil {
			s.RecordsAffected, s.Error = res.UpdateMany(s.UpdateDoc)
		} else {
			s.RecordsAffected, s.Error = res.DeleteMany()
		}
	}
}

func afterDeleteCallback(s *Scope) {
	s.callHooks(HookAfterDelete, s.Metadata.Name)
}
