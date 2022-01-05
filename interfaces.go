package db

import "context"

type Adapter interface {
	Name() string
	Connect(context.Context, DataSource, Logger) (Client, error)
}

type Client interface {
	Name() string
	Logger() Logger
	Source() DataSource
	Disconnect(context.Context) error
	StartTransaction() (Tx, error)
	WithTransaction(func(Tx) error) error
	Model(Metadata) Collection
}

type Collection interface {
	Name() string
	Metadata() Metadata
	Session() *Connection
	InsertOne(interface{}, ...*InsertOptions) (InsertOneResult, error)
	InsertMany(interface{}, ...*InsertOptions) (InsertManyResult, error)
	Find(...interface{}) Result
}

type Result interface {
	And(...Conditional) Result
	Or(...Conditional) Result
	Project(...string) Result
	One(dst interface{}) error
	All(dst interface{}) error
	Cursor() (Cursor, error)
	OrderBy(...string) Result
	Count() (int, error)
	Paginate(uint) Result
	Page(uint) Result
	TotalRecords() (int, error)
	Preload(string, ...*PreloadOptions) Result
	TotalPages() (int, error)
	UpdateOne(interface{}, ...*UpdateOptions) (int, error)
	UpdateMany(interface{}, ...*UpdateOptions) (int, error)
	Unscoped() Result
	DeleteOne(...*DeleteOptions) (int, error)
	DeleteMany(...*DeleteOptions) (int, error)
}

type Tx interface {
	Model(string) Collection
	Commit() error
	Rollback() error
}

type Cursor interface {
	HasNext() bool
	Next(dst interface{}) error
	Close() error
}

type InsertOneResult interface {
	StringID() string
	IntID() int
}

type InsertManyResult interface {
	StringIDs() []string
	IntIDs() []int
}

type QueryResult interface {
	One(dst interface{}) error
	All(dst interface{}) error
	Cursor() (Cursor, error)
}

type ExecResult interface {
	OK() bool
	RecordsAffected() int
}
