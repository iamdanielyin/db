package db

import (
	"github.com/buger/jsonparser"
	"github.com/gobwas/glob"
	"github.com/yuyitech/structs"
	"reflect"
	"strings"
	"sync"
)

const (
	HookBeforeSave   = "beforeSave"
	HookBeforeCreate = "beforeCreate"
	HookAfterCreate  = "afterCreate"
	HookAfterSave    = "afterSave"
	HookBeforeUpdate = "beforeUpdate"
	HookAfterUpdate  = "afterUpdate"
	HookBeforeQuery  = "beforeQuery"
	HookAfterQuery   = "afterQuery"
	HookBeforeDelete = "beforeDelete"
	HookAfterDelete  = "afterDelete"
)

const (
	HookFieldOperatorOr  = "OR"
	HookFieldOperatorAnd = "AND"
)

var AllHooks = []string{
	HookBeforeSave,
	HookBeforeCreate,
	HookAfterCreate,
	HookAfterSave,
	HookBeforeUpdate,
	HookAfterUpdate,
	HookBeforeQuery,
	HookAfterQuery,
	HookBeforeDelete,
	HookAfterDelete,
}

var (
	metadataHookMap   = make(map[string]MetadataHooks)
	metadataHookMapMu sync.RWMutex
)

type MetadataHooks map[string][]*MetadataHook

type MetadataHook struct {
	Pattern       string
	Action        string
	Fields        []string
	FieldOperator string
	Fn            func(*Scope)
}

func RegisterMiddleware(pattern string, fn func(*Scope)) error {
	metadataHookMapMu.Lock()
	defer metadataHookMapMu.Unlock()

	pattern = strings.TrimSpace(pattern)
	if pattern == "" || fn == nil {
		return nil
	}
	split := strings.Split(pattern, ":")
	if len(split) < 2 {
		return Errorf(`invalid middleware pattern: %s`, pattern)
	}
	split[0] = strings.TrimSpace(split[0])
	split[1] = strings.ToUpper(strings.TrimSpace(split[1]))
	var matchActions []string
	for _, name := range AllHooks {
		un := strings.ToUpper(name)
		g := glob.MustCompile(split[1])
		if g.Match(un) {
			matchActions = append(matchActions, name)
		}
	}
	if len(matchActions) == 0 {
		return Errorf(`invalid middleware pattern: %s`, pattern)
	}

	var (
		rule          = split[0]
		fields        []string
		fieldOperator string
	)
	if len(split) > 2 {
		split[2] = strings.TrimSpace(split[2])
		if idx := strings.Index(split[2], "|"); idx >= 0 {
			fields = strings.Split(split[2], "|")
			fieldOperator = HookFieldOperatorOr
		} else {
			fields = strings.Split(split[2], ",")
			fieldOperator = HookFieldOperatorAnd
		}
		for i, item := range fields {
			item = strings.TrimSpace(item)
			fields[i] = item
		}
	}

	for _, action := range matchActions {
		hook := &MetadataHook{
			Pattern:       rule,
			Action:        action,
			Fn:            fn,
			Fields:        fields,
			FieldOperator: fieldOperator,
		}
		metadataMapMu.RLock()
		for _, v := range metadataMap {
			g := glob.MustCompile(hook.Pattern)
			if g.Match(v.Name) {
				if metadataHookMap[v.Name] == nil {
					metadataHookMap[v.Name] = make(MetadataHooks)
				}
				metadataHookMap[v.Name][hook.Action] = append(metadataHookMap[v.Name][hook.Action], hook)
				continue
			}
		}
		metadataMapMu.RUnlock()
	}
	return nil
}

func testFieldsHook(hook *MetadataHook, action string, value interface{}) bool {
	if IsNil(value) {
		return false
	}
	fieldMap := make(map[string]bool)
	for _, field := range hook.Fields {
		fieldMap[field] = true
	}
	if action == ActionInsertOne || action == ActionInsertMany ||
		action == ActionUpdateOne || action == ActionUpdateMany {
		reflectValue := reflect.Indirect(reflect.ValueOf(value))
		switch reflectValue.Kind() {
		case reflect.Struct:
			s := structs.New(value)
			exists := make(map[string]bool)
			for _, name := range hook.Fields {
				field, ok := s.FieldOk(name)
				if !ok || field.IsZero() {
					if hook.FieldOperator == HookFieldOperatorOr {
						continue
					}
					return false
				}
				if hook.FieldOperator == HookFieldOperatorOr {
					return true
				} else {
					exists[name] = true
					if len(exists) == len(hook.Fields) {
						return true
					}
				}
			}
		case reflect.Array, reflect.Slice:
			for i := 0; i < reflectValue.Len(); i++ {
				item := reflectValue.Index(i).Interface()
				if ok := testFieldsHook(hook, action, item); ok {
					return true
				}
			}
		default:
			data, err := JSONMarshal(value)
			if err != nil {
				return false
			}
			exists := make(map[string]bool)
			for _, name := range hook.Fields {
				if _, _, _, err := jsonparser.Get(data, name); err != nil {
					if hook.FieldOperator == HookFieldOperatorOr {
						continue
					}
					return false
				}
				if hook.FieldOperator == HookFieldOperatorOr {
					return true
				} else {
					exists[name] = true
					if len(exists) == len(hook.Fields) {
						return true
					}
				}
			}
		}
	}
	return false
}
