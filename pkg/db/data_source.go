package db

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/yuyitech/db/pkg/cache"
	"sync"
)

var dsCache *safeDataSourcesCache

func init() {
	dsCache = &safeDataSourcesCache{
		l:             new(sync.RWMutex),
		dataSourceMap: make(map[string]*DataSource),
	}
}

type DataSource struct {
	db    Database
	cache cache.ICache

	Name      string `json:"name"`
	Adapter   string `json:"adapter"`
	DSN       string `json:"dsn"`
	IsDefault bool   `json:"is_default"`
}

type safeDataSourcesCache struct {
	l                 *sync.RWMutex
	dataSourceMap     map[string]*DataSource
	defaultDataSource *DataSource
}

func Connect(dataSource *DataSource) error {
	dsCache.l.Lock()
	defer dsCache.l.Unlock()

	msg := "add data source failed"
	if _, err := govalidator.ValidateStruct(*dataSource); err != nil {
		return errors.Wrap(err, msg)
	}
	if _, ok := dsCache.dataSourceMap[dataSource.Name]; ok {
		return errors.Wrap(fmt.Errorf("db.AddDataSource() called twice for data source: %s", dataSource.Name), msg)
	}

	adapter := adapters[dataSource.Adapter]
	if adapter == nil {
		return errors.Wrap(fmt.Errorf("unsupported adapter: %s", dataSource.Adapter), msg)
	}

	db, err := adapter.Open(dataSource)
	if err != nil {
		return errors.Wrap(err, msg)
	}
	dataSource.db = db

	dsCache.dataSourceMap[dataSource.Name] = dataSource
	if dataSource.IsDefault {
		dsCache.defaultDataSource = dataSource
	}

	return nil
}

func Disconnect(name string) error {
	dsCache.l.Lock()
	defer dsCache.l.Unlock()

	if v, has := dsCache.dataSourceMap[name]; has {
		delete(dsCache.dataSourceMap, name)
		return v.db.Close()
	}

	return nil
}

func DisconnectAll() (err error) {
	dsCache.l.Lock()
	defer dsCache.l.Unlock()

	for _, item := range dsCache.dataSourceMap {
		if v := item.db.Close(); v != nil && err == nil {
			err = v
		}
		delete(dsCache.dataSourceMap, item.Name)
	}

	return
}

func Session(name ...string) Database {
	dsCache.l.RLock()
	defer dsCache.l.RUnlock()

	var n string
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}

	var dd Database
	if n != "" {
		if v, has := dsCache.dataSourceMap[n]; has && v != nil && v.db != nil {
			dd = v.db
		}
	} else {
		if v := dsCache.defaultDataSource; v != nil {
			dd = v.db
		}
	}

	return &database{target: dd}
}
