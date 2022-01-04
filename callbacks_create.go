package db

func registerCreateCallbacks(callbacks *clientWrapper) *clientWrapper {
	processor := callbacks.Create()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:before_create", beforeCreateCallback)
	processor.Register("db:create", createCallback)
	processor.Register("db:after_create", afterCreateCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
	return callbacks
}

func beforeCreateCallback(s *Scope) {
	s.callHooks(HookBeforeSave, s.Metadata.Name)
	s.callHooks(HookBeforeCreate, s.Metadata.Name)
}

func createCallback(s *Scope) {
	if s.HasError() {
		return
	}

	switch s.Action {
	case ActionInsertOne:
		s.InsertOneResult, s.Error = s.Coll.InsertOne(s.InsertOneDoc)
	case ActionInsertMany:
		s.InsertManyResult, s.Error = s.Coll.InsertMany(s.InsertManyDocs)
	}
}

func afterCreateCallback(s *Scope) {
	s.callHooks(HookAfterCreate, s.Metadata.Name)
	s.callHooks(HookAfterSave, s.Metadata.Name)
}
