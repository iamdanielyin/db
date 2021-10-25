package db

import (
	"github.com/gobwas/glob"
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
	HookBeforeFind   = "beforeFind"
	HookAfterFind    = "afterFind"
	HookBeforeDelete = "beforeDelete"
	HookAfterDelete  = "afterDelete"
)

const (
	HookFieldOperatorOr  = "OR"
	HookFieldOperatorAnd = "AND"
)

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
	var matchAction string
	for _, action := range []string{
		HookBeforeSave,
		HookBeforeCreate,
		HookAfterCreate,
		HookAfterSave,
		HookBeforeUpdate,
		HookAfterUpdate,
		HookBeforeFind,
		HookAfterFind,
		HookBeforeDelete,
		HookAfterDelete,
	} {
		if strings.ToUpper(action) == split[1] {
			matchAction = action
			break
		}
	}
	if matchAction == "" {
		return Errorf(`invalid middleware pattern: %s`, pattern)
	}

	hook := &MetadataHook{
		Pattern: split[0],
		Action:  matchAction,
		Fn:      fn,
	}
	if len(split) > 2 {
		split[2] = strings.TrimSpace(split[2])
		if idx := strings.Index(split[2], ","); idx >= 0 {
			hook.Fields = strings.Split(split[2], ",")
			hook.FieldOperator = HookFieldOperatorAnd
		} else {
			hook.Fields = strings.Split(split[2], "|")
			hook.FieldOperator = HookFieldOperatorOr
		}
		for i, item := range hook.Fields {
			item = strings.TrimSpace(item)
			hook.Fields[i] = item
		}
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
	return nil
}

func callMetadataHooks(name, kind string, scope *Scope) {
	metadataHookMapMu.RLock()
	defer metadataHookMapMu.RUnlock()

	var hooks []*MetadataHook
	if v := metadataHookMap[name]; v != nil {
		hooks = v[kind]
	}

	for _, hook := range hooks {
		if len(hook.Fields) > 0 {
			// 字段中间件
		} else {
			// 元数据中间件
			hook.Fn(scope)
		}
	}
}
