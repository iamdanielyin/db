package db

func init() {
	DefaultAdapterMiddleware.Create().Register("db:begin_transaction", beginTransactionMiddleware)
	DefaultAdapterMiddleware.Create().Register("db:before_create", beforeCreateMiddleware)
	DefaultAdapterMiddleware.Create().Register("db:create", createMiddleware)
	DefaultAdapterMiddleware.Create().Register("db:after_create", afterCreateMiddleware)
	DefaultAdapterMiddleware.Create().Register("db:commit_or_rollback_transaction", commitOrRollbackTransactionMiddleware)
}

func beforeCreateMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("BeforeSave")
	}
	if !scope.HasError() {
		scope.CallHooks("BeforeCreate")
	}
}

func createMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.OutputValue, scope.RecordsAffected, scope.Error = scope.model.Create(scope.InputValue)
	}
}

func afterCreateMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("AfterCreate")
	}
	if !scope.HasError() {
		scope.CallHooks("AfterSave")
	}
}
