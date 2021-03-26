package db

func init() {
	DefaultAdapterMiddleware.Update().Register("db:begin_transaction", beginTransactionMiddleware)
	DefaultAdapterMiddleware.Update().Register("db:before_update", beforeUpdateMiddleware)
	DefaultAdapterMiddleware.Update().Register("db:update", updateMiddleware)
	DefaultAdapterMiddleware.Update().Register("db:after_update", afterUpdateMiddleware)
	DefaultAdapterMiddleware.Update().Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionMiddleware)
}

func beforeUpdateMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("BeforeSave")
	}
	if !scope.HasError() {
		scope.CallHooks("BeforeUpdate")
	}
}

func updateMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.RecordsAffected, scope.Error = scope.search.Update(scope.InputValue)
	}
}

func afterUpdateMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("AfterUpdate")
	}
	if !scope.HasError() {
		scope.CallHooks("AfterSave")
	}
}
