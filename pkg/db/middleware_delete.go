package db

func init() {
	DefaultAdapterMiddleware.Delete().Register("db:begin_transaction", beginTransactionMiddleware)
	DefaultAdapterMiddleware.Delete().Register("db:before_delete", beforeDeleteMiddleware)
	DefaultAdapterMiddleware.Delete().Register("db:delete", deleteMiddleware)
	DefaultAdapterMiddleware.Delete().Register("db:after_delete", afterDeleteMiddleware)
	DefaultAdapterMiddleware.Delete().Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionMiddleware)
}

func beforeDeleteMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("BeforeDelete")
	}
}

func deleteMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.RecordsAffected, scope.Error = scope.search.Delete()
	}
}

func afterDeleteMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("AfterDelete")
	}
}
