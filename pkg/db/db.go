package db

import (
	"github.com/yuyitech/db/pkg/schema"
	"sync"
)

var (
	adapters   = make(map[string]Database)
	adaptersMu sync.RWMutex
)

type BaseDatabase interface {
	Model(string) Collection
	Name() string
	Query(string, ...interface{}) Query
	Exec(string, ...interface{}) (interface{}, uint64, error)
	DataSource() *DataSource
}

type Database interface {
	BaseDatabase
	DriverName() string
	Open(*DataSource) (Database, error)
	Close() error
	NativeCollectionNames() ([]string, error)
	NativeCollectionMetadata() ([]schema.Metadata, error)
	BeginTx() (Tx, error)
}

type Iterator interface {
	Next(ptrToStruct interface{}) bool
	Err() error
	Close() error
}

type Query interface {
	Iterator() (Iterator, error)
	One(ptrToStruct interface{}) error
	All(sliceOfStruct interface{}) error
}

type Collection interface {
	Name() string
	Database() Database
	Metadata() schema.Metadata
	Create(interface{}) (interface{}, uint64, error)
	Find(...interface{}) FindResult
	Middleware() *AdapterMiddleware
}

type PopulateOptions struct {
	Select []string
	Model  Collection
	Match  interface{}
}

type FindResult interface {
	Query

	Page(uint) FindResult
	Size(uint) FindResult
	Order(...string) FindResult
	Select(...string) FindResult
	Where(interface{}) FindResult
	And(...Cond) FindResult
	Or(...Cond) FindResult
	Populate(string, ...*PopulateOptions) FindResult
	TotalPages() (uint, error)
	TotalRecords() (uint64, error)
	Count() (uint64, error)
	Delete() (uint64, error)
	Update(interface{}) (uint64, error)
}

type Tx interface {
	BaseDatabase

	Rollback() error
	Commit() error
}

func RegisterAdapter(adapter Database) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()

	name := adapter.DriverName()

	if name == "" {
		panic(`Missing adapter name`)
	}
	if _, ok := adapters[name]; ok {
		panic(`db.RegisterAdapter() called twice for adapter: ` + name)
	}
	adapters[name] = adapter
}
