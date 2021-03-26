package db

const (
	queryActionOne      = "one"
	queryActionAll      = "all"
	queryActionCount    = "count"
	queryActionIterator = "iterator"
)

func init() {
	DefaultAdapterMiddleware.Query().Register("db:before_query", beforeQueryMiddleware)
	DefaultAdapterMiddleware.Query().Register("db:query", queryMiddleware)
	DefaultAdapterMiddleware.Query().Register("db:after_query", afterQueryMiddleware)
}

func beforeQueryMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("BeforeFind")
	}
}

func queryMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		switch scope.queryAction {
		case queryActionOne:
			scope.Error = scope.search.One(scope.dst)
		case queryActionAll:
			scope.Error = scope.search.All(scope.dst)
		case queryActionCount:
			scope.RecordsAffected, scope.Error = scope.search.Count()
		case queryActionIterator:
			scope.Iterator, scope.Error = scope.search.Iterator()
		}
	}
}

func afterQueryMiddleware(scope *AdapterMiddlewareScope) {
	if !scope.HasError() {
		scope.CallHooks("AfterFind")
	}
}
