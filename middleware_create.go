package db

func registerCreateCallbacks(callbacks *callbacks) {
	processor := callbacks.Create()
	processor.Register("db:begin_transaction", beginTransactionCallback)
	processor.Register("db:before_create", beforeCreateCallback)
	processor.Register("db:save_before_associations", saveBeforeAssociationsCallback)
	processor.Register("db:create", createCallback)
	processor.Register("db:save_after_associations", saveAfterAssociationsCallback)
	processor.Register("db:after_create", afterCreateCallback)
	processor.Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionCallback)
}

func beginTransactionCallback(s *Scope) {

}

func beforeCreateCallback(s *Scope) {

}

func saveBeforeAssociationsCallback(s *Scope) {

}

func createCallback(s *Scope) {

}

func saveAfterAssociationsCallback(s *Scope) {

}

func afterCreateCallback(s *Scope) {

}

func commitOrRollbackTransactionCallback(s *Scope) {

}
