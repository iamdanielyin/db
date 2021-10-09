package db

import (
	"github.com/asaskevich/govalidator"
	"strings"
	"sync"
)

var (
	connMap   = make(map[string]IConnection)
	connMapMu sync.RWMutex
)

type DataSource struct {
	Name    string `valid:"required,!empty"`
	Adapter string `valid:"required,!empty"`
	URI     string `valid:"required,!empty"`
}

func Connect(source DataSource) (IConnection, error) {
	connMapMu.Lock()
	defer connMapMu.Unlock()

	if _, err := govalidator.ValidateStruct(&source); err != nil {
		return nil, Errorf(err.Error())
	}

	adapterMapMu.RLock()
	adapter := adapterMap[source.Adapter]
	adapterMapMu.RUnlock()

	if adapter == nil {
		return nil, Errorf(`unregistered adapter "%s"`, source.Adapter)
	}

	if _, has := connMap[source.Name]; has {
		return nil, Errorf(`data source name already exists "%s"`, source.Name)
	}

	conn, err := adapter.Connect(source)
	if err != nil {
		return nil, err
	}
	connMap[source.Name] = conn
	return conn, nil
}

func Disconnect(name ...string) error {
	connMapMu.Lock()
	defer connMapMu.Unlock()

	for _, item := range name {
		item = strings.TrimSpace(item)
		if v, has := connMap[item]; has && v != nil {
			if err := v.Disconnect(); err != nil {
				return err
			}
		}
	}
	return nil
}

func Session(name string) IConnection {
	connMapMu.RLock()
	defer connMapMu.RUnlock()

	conn, has := connMap[name]
	if !has || conn == nil {
		panic(Errorf(`missing session: %s`, name))
	}
	return conn
}
