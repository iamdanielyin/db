package db

import "github.com/yuyitech/db/pkg/schema"

type AdapterMiddlewareScope struct {
	model    Collection
	tx       Tx
	skipLeft bool

	InputValue      interface{}
	OutputValue     interface{}
	Error           error
	RecordsAffected uint64
	Iterator        Iterator

	dst         interface{}
	queryAction interface{}
	search      FindResult
}

func (scope *AdapterMiddlewareScope) SkipLeft() {
	scope.skipLeft = true
}

func (scope *AdapterMiddlewareScope) HasError() bool {
	return scope.Error != nil
}

func (scope *AdapterMiddlewareScope) FieldByName(name string) (field schema.Field, ok bool) {
	field, ok = scope.model.Metadata().Fields[name]
	return
}

func (scope *AdapterMiddlewareScope) CallHooks(hookName string) {
	if hooks := ModelHooks(scope.model.Name(), hookName); len(hooks) > 0 {
		for _, h := range hooks {
			err := h(scope)
			if err != nil {
				scope.Error = err
				break
			}
		}
	}
}

func (scope *AdapterMiddlewareScope) PrimaryFields() []schema.Field {
	meta := scope.model.Metadata()
	return (&meta).PrimaryFields()
}

func (scope *AdapterMiddlewareScope) PrimaryField() *schema.Field {
	if primaryFields := scope.PrimaryFields(); len(primaryFields) > 0 {
		if len(primaryFields) > 0 {
			return &(scope.PrimaryFields()[0])
		}
	}
	return nil
}

func (scope *AdapterMiddlewareScope) callMiddleware(fns []*func(s *AdapterMiddlewareScope)) *AdapterMiddlewareScope {
	defer func() {
		if err := recover(); err != nil {
			if scope.tx != nil {
				_ = scope.tx.Rollback()
			}
			panic(err)
		}
	}()
	for _, f := range fns {
		(*f)(scope)
		if scope.skipLeft {
			break
		}
	}
	return scope
}
