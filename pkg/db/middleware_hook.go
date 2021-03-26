package db

import (
	"fmt"
	"sync"
)

var (
	modelHooks   = make(map[string][]func(*AdapterMiddlewareScope) error)
	modelHooksMu sync.RWMutex
)

func RegisterModelHooks(modelName, hookName string, handlers ...func(*AdapterMiddlewareScope) error) {
	if modelName == "" || hookName == "" || len(handlers) == 0 {
		return
	}

	modelHooksMu.Lock()
	defer modelHooksMu.Unlock()

	key := fmt.Sprintf("%s.%s", modelName, hookName)
	hooks := modelHooks[key]
	hooks = append(hooks, handlers...)

	modelHooks[key] = hooks
}

func ModelHooks(modelName, hookName string) []func(*AdapterMiddlewareScope) error {
	if modelName == "" || hookName == "" {
		return nil
	}

	modelHooksMu.RLock()
	defer modelHooksMu.RUnlock()

	key := fmt.Sprintf("%s.%s", modelName, hookName)
	hooks := modelHooks[key]

	return hooks
}

func UnregisterModelHooks(modelName, hookName string) {
	if modelName == "" || hookName == "" {
		return
	}

	modelHooksMu.Lock()
	defer modelHooksMu.Unlock()

	key := fmt.Sprintf("%s.%s", modelName, hookName)

	delete(modelHooks, key)
}
