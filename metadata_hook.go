package db

import (
	"strings"
	"sync"
)

var (
	metadataHooks   = make([]*MetadataHook, 0)
	metadataHooksMu sync.RWMutex
)

type MetadataHook struct {
	Metadata string
	Fields   []string
	Fn       func(*Scope)
}

func RegisterMiddleware(pattern string, fn func(*Scope)) error {
	if pattern == "" || fn == nil {
		return nil
	}
	split := strings.Split(pattern, ":")
	if len(split) < 2 {
		return Errorf(`incorrect middleware pattern: %s`, pattern)
	}

	//for _, item := range split {
	//
	//}
	//hook := &MetadataHook{
	//	Fields: pattern,
	//	Fn:     fn,
	//}

	return nil
}
