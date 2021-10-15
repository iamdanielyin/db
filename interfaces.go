package db

import "context"

type Adapter interface {
	Name() string
	Connect(context.Context, DataSource) (Client, error)
}

type Client interface {
	Name() string
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
	InsertOne(interface{}) (InsertOneResult, error)
	InsertMany(interface{}) (InsertManyResult, error)
	Find(...interface{}) Result
}

type Result interface {
	And(...interface{}) Result
	Or(...interface{}) Result
	Project(...string) Result
	QueryResult
	OrderBy(...string) Result
	Count() (int, error)
	Paginate(uint) Result
	Page(uint) Result
	TotalRecords() (int, error)
	TotalPages() (int, error)
	UpdateOne(interface{}) (int, error)
	UpdateMany(interface{}) (int, error)
	Unscoped() Result
	DeleteOne() (int, error)
	DeleteMany() (int, error)
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
