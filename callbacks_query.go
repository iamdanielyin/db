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
	if len(scope.PreloadsOptions) == 0 {
		return
	}
	for _, preloadItem := range scope.PreloadsOptions {
		preloadFields := strings.Split(strings.TrimSpace(preloadItem.Path), ".")
		for _, preloadFieldName := range preloadFields {
			preloadField, has := scope.Metadata.FieldByName(preloadFieldName)
			if !has || !preloadField.IsRef {
				continue
			}
			rel := preloadField.Relationship
			switch rel.Type {
			case RelationshipHasOne:
			case RelationshipHasMany:
			case RelationshipRefOne:
			case RelationshipRefMany:
			}
		}

		dstModel := scope.Session.Model()
	}
	switch s.Action {
	case ActionQueryOne:

	case ActionQueryAll:

	case ActionQueryCursor:

	}
}

func preloadField(value interface{}, meta Metadata, opts *PreloadOptions, sess *Connection) error {
	var relationship Relationship
	if f, has := meta.FieldByName(opts.Path); has {
		relationship = f.Relationship
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
		if err := execPreload(relationship, sess, srcValue, targetValue); err != nil {
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
			if err := execPreload(relationship, sess, srcValue, targetValue); err != nil {
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

func execPreload(relationship Relationship, sess *Connection, srcValue interface{}, targetValue interface{}) error {
	if _, err := LookupMetadata(relationship.Metadata); err != nil {
		return err
	}
	refModel := sess.Model(relationship.Metadata)

	switch relationship.Type {
	case RelationshipHasOne, RelationshipRefOne:
		if err := refModel.Find(Cond{
			relationship.DstField: srcValue,
		}).One(targetValue); err != nil {
			return err
		}
	case RelationshipHasMany:
		if err := refModel.Find(Cond{
			relationship.DstField: srcValue,
		}).All(targetValue); err != nil {
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
