package db

import "sync"

var (
	adapterMap   = make(map[string]IAdapter)
	adapterMapMu sync.RWMutex
)

func RegisterAdapter(name string, adapter IAdapter) {
	adapterMapMu.Lock()
	defer adapterMapMu.Unlock()

	if name == "" {
		panic(Errorf(`missing adapter name`))
	}
	if _, ok := adapterMap[name]; ok {
		panic(Errorf(`db.RegisterAdapter() called twice for adapter: %s`, name))
	}
	adapterMap[name] = adapter
}
