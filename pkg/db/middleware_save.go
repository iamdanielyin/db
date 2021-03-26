package db

func beginTransactionMiddleware(scope *AdapterMiddlewareScope) {
	scope.tx, scope.Error = scope.model.Database().BeginTx()
	scope.model = scope.tx.Model(scope.model.Name())
}

func commitOrRollbackTransactionMiddleware(scope *AdapterMiddlewareScope) {
	if scope.tx != nil {
		if scope.HasError() {
			scope.Error = scope.tx.Rollback()
		} else {
			scope.Error = scope.tx.Commit()
		}
	}
}
