package db

type IAdapter interface {
	Connect(source DataSource) (IConnection, error)
	Disconnect() error
}

type IConnection interface {
	AdapterName() string
	DataSource() DataSource
	RegisterMetadata(Metadata) error
	Disconnect() error
}

type ISession interface {
}
