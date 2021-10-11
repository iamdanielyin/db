package db

type Collection struct {
	Metadata Metadata
	Session  *Connection
}

func (c *Collection) InsertOne(doc interface{}) (*InsertOneResult, error) {
	return nil, nil
}

func (c *Collection) InsertMany(docs interface{}) (*InsertManyResult, error) {
	return nil, nil
}

func (c *Collection) Find(...interface{}) Result {
	return nil
}

type Result interface {
	And(...interface{}) Result
	Or(...interface{}) Result
	Project(map[string]int) Result
	QueryResult
	OrderBy(...string) Result
	Count() (int, error)
	Paginate(uint) Result
	Page(uint) Result
	TotalRecords() (int, error)
	TotalPages() (int, error)
	UpdateOne(interface{}) (UpdateOneResult, error)
	UpdateMany(interface{}) (UpdateManyResult, error)
	Unscoped() Result
	DeleteOne(interface{}) (DeleteOneResult, error)
	DeleteMany(interface{}) (DeleteManyResult, error)
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

type InsertOneResult struct {
	result interface{}
}

func (i *InsertOneResult) StringID() string {
	return ""
}

func (i *InsertOneResult) IntID() int {
	return 0
}

type InsertManyResult interface {
	StringIDs() []string
	IntIDs() []int
}

type UpdateOneResult interface {
	ExecResult
}

type UpdateManyResult interface {
	ExecResult
}

type DeleteOneResult interface {
	ExecResult
}

type DeleteManyResult interface {
	ExecResult
}

type QueryResult interface {
	One(dst interface{}) error
	All(dst interface{}) error
	Cursor() Cursor
}

type ExecResult interface {
	OK() bool
	RecordsAffected() int
}
