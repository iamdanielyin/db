package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"time"
)

type mongoClient struct {
	adapter *mongoAdapter
	source  db.DataSource
	cs      connstring.ConnString
	client  *mongo.Client
}

func (c *mongoClient) Name() string {
	return c.source.Name
}

func (c *mongoClient) Source() db.DataSource {
	return c.source
}

func (c *mongoClient) Model(metadata db.Metadata) db.Collection {
	mdb := c.client.Database(c.cs.Database)
	coll := mdb.Collection(metadata.MustNativeName())
	return &mongoCollection{
		client: c,
		sess:   metadata.Session(),
		meta:   metadata,
		db:     mdb,
		coll:   coll,
	}
}

func (c *mongoClient) Disconnect(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}
