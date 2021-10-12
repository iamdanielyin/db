package db

import (
	"context"
	"github.com/asaskevich/govalidator"
	"strings"
	"sync"
)

var (
	connMap   = make(map[string]*Connection)
	connMapMu sync.RWMutex
)

type Connection struct {
	client Client
	//callbacks  *callbacks
	cacheStore *sync.Map
}

func (c *Connection) Client() Client {
	return c.client
}

func (c *Connection) Disconnect() error {
	return c.client.Disconnect(context.Background())
}

func (c *Connection) StartTransaction() (Tx, error) {
	return nil, nil
}

func (c *Connection) WithTransaction(fn func(Tx) error) (err error) {
	var tx Tx
	tx, err = c.StartTransaction()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			if e := tx.Rollback(); e != nil {
				err = Errorf("%v; %w", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(tx)
	return
}

func Connect(source DataSource) (*Connection, error) {
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

	client, err := adapter.Connect(context.Background(), source)
	if err != nil {
		return nil, err
	}
	conn := &Connection{
		client:     client,
		cacheStore: &sync.Map{},
	}
	connMap[source.Name] = conn
	return conn, nil
}

func Disconnect(names ...string) error {
	connMapMu.Lock()
	defer connMapMu.Unlock()

	if len(names) == 0 {
		for k, _ := range connMap {
			names = append(names, k)
		}
	}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if v, has := connMap[name]; has && v != nil {
			if err := v.Disconnect(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Connection) RegisterMetadata(meta Metadata) error {
	//TODO 待实现元数据注册
	return nil
}
