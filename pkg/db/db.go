package db

import (
	"sync"
)

var (
	adapters   = make(map[string]IDatabase)
	adaptersMu sync.RWMutex
)

type IBaseDatabase interface {
	Model(string) IModel
	Name() string
	Query(string, ...interface{}) IQuery
	Exec(string, ...interface{}) (interface{}, uint64, error)
	DataSource() *DataSource
}

type IDatabase interface {
	IBaseDatabase
	DriverName() string
	Open(*DataSource) (IDatabase, error)
	Close() error
	NativeCollectionNames() ([]string, error)
	NativeCollectionMetadata() ([]Metadata, error)
	BeginTx() (ITx, error)
}

type Iterator interface {
	Next(ptrToStruct interface{}) bool
	Err() error
	Close() error
}

type IQuery interface {
	Iterator() (Iterator, error)
	One(ptrToStruct interface{}) error
	All(sliceOfStruct interface{}) error
}

type IModel interface {
	Name() string
	Database() IDatabase
	Metadata() Metadata
	Create(interface{}) (interface{}, uint64, error)
	Find(...interface{}) IFindResult
	Middleware() *AdapterMiddleware
}

type PopulateOptions struct {
	Select []string
	Model  IModel
	Match  interface{}
}

type IFindResult interface {
	IQuery

	Page(uint) IFindResult
	Size(uint) IFindResult
	Order(...string) IFindResult
	Select(...string) IFindResult
	Where(interface{}) IFindResult
	And(...Cond) IFindResult
	Or(...Cond) IFindResult
	Populate(string, ...*PopulateOptions) IFindResult
	TotalPages() (uint, error)
	TotalRecords() (uint64, error)
	Count() (uint64, error)
	Delete() (uint64, error)
	Update(interface{}) (uint64, error)
}

type ITx interface {
	IBaseDatabase

	Rollback() error
	Commit() error
}

func RegisterAdapter(adapter IDatabase) {
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
