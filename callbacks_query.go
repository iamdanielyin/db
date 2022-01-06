package db

import (
	"github.com/iamdanielyin/structs"
	"log"
	"reflect"
	"strings"
)

func registerQueryCallbacks(callbacks *clientWrapper) *clientWrapper {
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
		log.Println("=======", s.Dest)
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
		preloadFields := strings.Split(preloadItem.Path, ".")
		meta := &scope.Metadata
		for _, preloadFieldName := range preloadFields {
			if meta == nil {
				break
			}
			field, has := meta.FieldByName(preloadFieldName)
			if !has {
				break
			}
			var opts *PreloadOptions
			if len(preloadFields) > 1 {
				opts = &PreloadOptions{
					Path: preloadFieldName,
				}
			} else {
				opts = &PreloadOptions{
					Path:     preloadFieldName,
					Select:   preloadItem.Select,
					OrderBys: preloadItem.OrderBys,
					Page:     preloadItem.Page,
					Size:     preloadItem.Size,
				}
			}
			if err := preloadField(scope.Dest, meta, opts, scope.Session); err != nil {
				scope.AddError(err)
				break
			}
			if v, err := LookupMetadata(field.Relationship.MetadataName); err != nil {
				scope.AddError(err)
				break
			} else {
				meta = &v
			}
		}
	}
}

func preloadField(value interface{}, meta *Metadata, opts *PreloadOptions, sess *Connection) error {
	if IsNil(value) || meta == nil || opts == nil || sess == nil {
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
		if f, ok := ss.FieldOk(relationship.SrcFieldName); !ok || f.IsZero() {
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
			if key == relationship.SrcFieldName {
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
	case reflect.Array, reflect.Slice:
		// TODO 联查字段为数组时
		// 内存组装 or [联表查询]？
	}
	return nil
}

func execPreload(relationship *Relationship, opts *PreloadOptions, sess *Connection, srcValue interface{}, targetValue interface{}) error {
	refModel := sess.Model(relationship.MetadataName)
	switch relationship.Type {
	case RelationshipHasOne, RelationshipRefOne, RelationshipHasMany:
		res := opts.SetResult(refModel.Find(Cond{
			relationship.DstFieldName: srcValue,
		}))
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
		if _, err := LookupMetadata(relationship.IntermediateMetadataName); err != nil {
			return err
		}
		intModel := sess.Model(relationship.IntermediateMetadataName)
		var intData []map[string]interface{}
		if err := intModel.Find(Cond{
			relationship.IntermediateSrcFieldName: srcValue,
		}).All(&intData); err != nil {
			return err
		}
		var dstIDs []interface{}
		for _, item := range intData {
			dstID := item[relationship.IntermediateDstFieldName]
			rv := reflect.ValueOf(dstID)
			zero := reflect.Zero(rv.Type()).Interface()
			current := rv.Interface()
			if !reflect.DeepEqual(current, zero) {
				dstIDs = append(dstIDs, dstID)
			}
		}
		if len(dstIDs) > 0 {
			res := opts.SetResult(refModel.Find(Cond{}.In(relationship.DstFieldName, dstIDs)))
			if err := res.All(targetValue); err != nil {
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
