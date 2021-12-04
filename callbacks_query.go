package db

import (
	"github.com/yuyitech/structs"
	"reflect"
	"strings"
)

func registerQueryCallbacks(callbacks *callbacks) *callbacks {
	processor := callbacks.Query()
	processor.Register("db:before_query", beforeQueryCallback)
	processor.Register("db:query", queryCallback)
	processor.Register("db:preload", preloadCallback)
	processor.Register("db:after_query", afterQueryCallback)
	return callbacks
}

func beforeQueryCallback(s *Scope) {
	s.callHooks(HookBeforeQuery, s.Metadata.Name)
}

func queryCallback(s *Scope) {
	if s.HasError() {
		return
	}

	res := s.buildQueryResult()

	switch s.Action {
	case ActionQueryOne:
		s.Error = res.One(s.Dest)
	case ActionQueryAll:
		s.Error = res.All(s.Dest)
	case ActionQueryCursor:
		s.Cursor, s.Error = res.Cursor()
	case ActionQueryCount:
		s.TotalRecords, s.Error = res.TotalRecords()
	case ActionQueryPage:
		s.TotalPages, s.Error = res.TotalPages()
	}
}

func preloadCallback(scope *Scope) {
	if len(scope.PreloadsOptions) == 0 || (scope.Action != ActionQueryOne && scope.Action != ActionQueryAll) {
		return
	}
	for _, preloadItem := range scope.PreloadsOptions {
		preloadFields := strings.Split(strings.TrimSpace(preloadItem.Path), ".")
		meta := scope.Metadata
		for _, preloadFieldName := range preloadFields {
			f, has := meta.FieldByName(preloadFieldName)
			if !has {
				break
			}
			if err := preloadField(scope.Dest, meta, &PreloadOptions{
				Path:     preloadFieldName,
				Select:   preloadItem.Select,
				OrderBys: preloadItem.OrderBys,
				Page:     preloadItem.Page,
				Size:     preloadItem.Size,
			}, scope.Session); err != nil {
				scope.AddError(err)
				break
			}
			//f.Relationship
			// TODO 链式下一个
		}
	}
}

func preloadField(value interface{}, meta Metadata, opts *PreloadOptions, sess *Connection) error {
	// TODO value会是数组
	if IsNil(value) || sess == nil {
		return nil
	}
	var relationship *Relationship
	if f, has := meta.FieldByName(opts.Path); has {
		relationship = &f.Relationship
	} else {
		return Errorf("preload field does not exist: %s", opts.Path)
	}
	if relationship.Type == "" {
		return Errorf("undefined relationship: %s", opts.Path)
	}
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	switch indirectValue.Kind() {
	case reflect.Struct:
		var srcValue interface{}
		ss := structs.New(value)
		if f, ok := ss.FieldOk(relationship.SrcField); !ok || f.IsZero() {
			return nil
		} else {
			srcValue = f.Value()
		}
		var targetField *structs.Field
		var targetValue interface{}
		if f, ok := ss.FieldOk(opts.Path); !ok {
			return nil
		} else {
			targetField = f
			targetValue = reflect.New(reflect.PtrTo(f.ReflectField().Type)).Interface()
		}
		if err := execPreload(relationship, opts, sess, srcValue, targetValue); err != nil {
			return err
		}
		if targetField.Kind() == reflect.Ptr {
			targetField.Set(targetValue)
		} else {
			targetField.Set(reflect.ValueOf(targetValue).Elem().Interface())
		}
	case reflect.Map:
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		var srcValue interface{}
		for _, k := range indirectValue.MapKeys() {
			key := k.Interface().(string)
			if key == relationship.SrcField {
				srcValue = indirectValue.MapIndex(k).Interface()
				break
			}
		}
		if srcValue != nil {
			var targetValue interface{}
			if err := execPreload(relationship, opts, sess, srcValue, targetValue); err != nil {
				return err
			}
			for _, k := range indirectValue.MapKeys() {
				key := k.Interface().(string)
				if key == opts.Path {
					indirectValue.MapIndex(k).Set(reflect.ValueOf(targetValue))
					break
				}
			}
		}

	}
	return nil
}

func execPreload(relationship *Relationship, opts *PreloadOptions, sess *Connection, srcValue interface{}, targetValue interface{}) error {
	if _, err := LookupMetadata(relationship.Metadata); err != nil {
		return err
	}
	refModel := sess.Model(relationship.Metadata)

	switch relationship.Type {
	case RelationshipHasOne, RelationshipRefOne, RelationshipHasMany:
		res := refModel.Find(Cond{
			relationship.DstField: srcValue,
		})
		if !IsNil(opts.Match) {
			res.And(opts.Match)
		}
		if len(opts.Select) > 0 {
			res.Project(opts.Select...)
		}
		if len(opts.OrderBys) > 0 {
			res.OrderBy(opts.OrderBys...)
		}
		if opts.Size > 0 {
			res.Paginate(opts.Size)
		}
		if opts.Page > 0 {
			res.Page(opts.Page)
		}
		var err error
		if relationship.Type == RelationshipHasMany {
			err = res.All(targetValue)
		} else {
			err = res.One(targetValue)
		}
		if err != nil {
			return err
		}
	case RelationshipRefMany:
		if _, err := LookupMetadata(relationship.IntMeta); err != nil {
			return err
		}
		intModel := sess.Model(relationship.IntMeta)
		var intData []map[string]interface{}
		if err := intModel.Find(Cond{
			relationship.IntSrcField: srcValue,
		}).All(&intData); err != nil {
			return err
		}
		var dstIDs []interface{}
		for _, item := range intData {
			dstID := item[relationship.IntDstField]
			rv := reflect.ValueOf(dstID)
			zero := reflect.Zero(rv.Type()).Interface()
			current := rv.Interface()
			if !reflect.DeepEqual(current, zero) {
				dstIDs = append(dstIDs, dstID)
			}
		}
		if len(dstIDs) > 0 {
			if err := refModel.Find(Cond{}.In(relationship.DstField, dstIDs)).All(targetValue); err != nil {
				return err
			}
		}
	default:
		return Errorf("unsupported relationship type: %s", relationship.Type)
	}
	return nil
}

func afterQueryCallback(s *Scope) {
	s.callHooks(HookAfterQuery, s.Metadata.Name)
}
