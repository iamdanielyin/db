package mongo

import (
	"github.com/yuyitech/db"
)

type mongoConnection struct {
	adapter *mongoAdapter
}

func (c *mongoConnection) AdapterName() string {
	return Adapter
}

func (c *mongoConnection) DataSource() db.DataSource {
	return c.adapter.source
}

func (c *mongoConnection) RegisterMetadata(meta db.Metadata) error {
	return db.RegisterMetadata(c.adapter.source.Name, meta)
}

func (c *mongoConnection) Disconnect() error {
	return c.adapter.Disconnect()
}
