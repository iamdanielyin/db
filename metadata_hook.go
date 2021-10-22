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

var (
	metadataHooks   = make([]*MetadataHook, 0)
	metadataHooksMu sync.RWMutex
)

type MetadataHook struct {
	Pattern string
	Action  string
	Fields  string
	Fn      func(*Scope)
}

func RegisterMiddleware(pattern string, fn func(*Scope)) error {
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
		hook.Fields = strings.TrimSpace(split[2])
	}

	metadataHooksMu.Lock()
	metadataHooks = append(metadataHooks, hook)
	metadataHooksMu.Unlock()

	matchMetadataHooks()
	return nil
}

func matchMetadataHooks(names ...string) {
	metadataMapMu.Lock()
	metadataHooksMu.Lock()
	defer func() {
		metadataMapMu.Unlock()
		metadataHooksMu.Unlock()
	}()

	var nameMap = make(map[string]bool)
	for _, k := range names {
		nameMap[k] = true
	}

	for k, v := range metadataMap {
		if len(names) > 0 && !nameMap[k] {
			continue
		}
		for _, hook := range metadataHooks {
			g := glob.MustCompile(hook.Pattern)
			if g.Match(v.Name) {
				v.hooks = append(v.hooks, hook)
				continue
			}
		}
		metadataMap[k] = v
	}
}
