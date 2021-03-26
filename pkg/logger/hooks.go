package logger

import (
	"github.com/sirupsen/logrus"
	"sync"
)

func init() {
	logrus.AddHook(&simpleHooks{})
}

type HookFn func(*logrus.Entry) bool

var (
	hookFns   []HookFn
	hookFnsMu sync.RWMutex
)

func AddHook(fn ...HookFn) {
	hookFnsMu.Lock()
	defer hookFnsMu.Unlock()

	hookFns = append(hookFns, fn...)
}

type simpleHooks struct{}

func (h *simpleHooks) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *simpleHooks) Fire(entry *logrus.Entry) error {
	hookFnsMu.RLock()
	defer hookFnsMu.RUnlock()

	for _, fn := range hookFns {
		if skip := fn(entry); skip {
			return nil
		}
	}
	return nil
}
